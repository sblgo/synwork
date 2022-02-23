package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	version "github.com/hashicorp/go-version"
	"sbl.systems/go/synwork/synwork/processor/cfg"
)

const (
	pluginFileNamePatternString = `synwork-processor-(\w[\w\d]*)_([^_]+)_([\w\d]+_[\w\d]+)`
)

var (
	pluginFileNamePattern *regexp.Regexp
)

type PluginKey struct {
	Hostname      string
	Namespace     string
	Name          string
	Version       string
	OsArch        string
	Ext           string
	ParsedVersion *version.Version
}
type PluginSource struct {
	PluginKey
	Source            string
	VersionSelector   string
	Config            *cfg.Config
	PluginProgram     string
	VersionConstraint version.Constraints
	allVersions       []*PluginReference
}

type PluginReference struct {
	PluginKey
	loadFunction func(ctx context.Context) ([]byte, error)
}

type pluginKeySorter struct {
	keys []*PluginReference
}

func (s *pluginKeySorter) Len() int {
	return len(s.keys)
}

func (s *pluginKeySorter) Swap(i, j int) {
	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
}

func (s *pluginKeySorter) Less(i, j int) bool {
	if s.keys[i].ParsedVersion.Equal(s.keys[i].ParsedVersion) {
		if s.keys[i].loadFunction != nil && s.keys[j].loadFunction == nil {
			return false
		} else if s.keys[i].loadFunction == nil && s.keys[j].loadFunction != nil {
			return true
		} else {
			return false
		}
	}
	return !s.keys[i].ParsedVersion.GreaterThanOrEqual(s.keys[j].ParsedVersion)
}

func init() {
	pluginFileNamePattern = regexp.MustCompile(`^` + pluginFileNamePatternString + `$`)
}

func NewPluginKeyFromFile(Hostname, Namespace string, f fs.FileInfo) (*PluginKey, error) {
	matches := pluginFileNamePattern.FindStringSubmatch(f.Name())
	if matches == nil {
		return nil, fmt.Errorf("malformed filename: %s", f.Name())
	}
	key := &PluginKey{
		Hostname:  Hostname,
		Namespace: Namespace,
		Name:      matches[1],
		Version:   matches[2],
		OsArch:    matches[3],
	}
	if pVersion, err := version.NewVersion(key.Version); err != nil {
		return nil, fmt.Errorf("malformed filename (version): %s", f.Name())
	} else {
		key.ParsedVersion = pVersion
	}

	return key, nil
}

func (ps *PluginSource) ListLocalPluginVersions() ([]*PluginReference, error) {
	keys := []*PluginReference{}
	filenamePattern := regexp.MustCompile(fmt.Sprintf("^synwork-processor-%s%s$", ps.Name, ps.Ext))
	verDir := filepath.Join(ps.Config.PluginDir, ps.Hostname, ps.Namespace, ps.Name)
	versions, err := os.ReadDir(verDir)
	if err != nil {
		return nil, nil
	}
	for _, vers := range versions {
		if parsedVersion, err := version.NewVersion(vers.Name()); err == nil {
			if ps.VersionConstraint.Check(parsedVersion) {
				archDir := filepath.Join(verDir, vers.Name())
				if archives, err := os.ReadDir(archDir); err == nil {
					for _, archive := range archives {
						pluginDir := filepath.Join(archDir, archive.Name())
						if files, err := os.ReadDir(pluginDir); err == nil {
							for _, f := range files {

								if filenamePattern.MatchString(f.Name()) {
									key := &PluginReference{
										PluginKey: PluginKey{
											Hostname:      ps.Hostname,
											Namespace:     ps.Namespace,
											Name:          ps.Name,
											Version:       vers.Name(),
											ParsedVersion: parsedVersion,
											OsArch:        archive.Name(),
											Ext:           ps.Ext,
										},
									}
									keys = append(keys, key)
								}
							}
						}

					}
				}
			}
		}
	}
	return keys, nil
}

func (ps *PluginSource) verifyAndLoadPlugin() error {
	if err := ps.evalRemotePluginProgram(); err != nil {
		return err
	}
	if err := ps.evalLocalPluginProgram(); err != nil {
		return err
	}

	return nil
}

func (ps *PluginSource) verifyPlugin() error {
	if err := ps.evalLocalPluginProgram(); err != nil {
		return err
	}

	return nil
}

func NewPluginSourceFromSource(c *cfg.Config, s string) (*PluginSource, error) {
	pr := &PluginSource{
		Source: s,
		Config: c,
		PluginKey: PluginKey{
			Ext:       c.ProgramExt,
			Namespace: INTERN_PROVIDER,
			Hostname:  INTERN_HOSTNAME,
			OsArch:    c.OsArch,
		},
	}

	parts := strings.Split(pr.Source, "/")
	switch len(parts) {
	case 1:
		pr.Name = parts[0]
	case 2:
		pr.Namespace = parts[0]
		pr.Name = parts[1]
	case 3:
		pr.Hostname = parts[0]
		pr.Namespace = parts[1]
		pr.Name = parts[2]
	default:
		return nil, fmt.Errorf("malformed source description %s", s)
	}

	return pr, nil
}

func (pr *PluginSource) selectPluginProgram() error {
	parts := append(filepath.SplitList(pr.Config.PluginDir), pr.Hostname, pr.Namespace, pr.Name)
	progName := func(ver string) string {
		progParts := append(parts, ver, pr.Config.OsArch, fmt.Sprintf("synwork-processor-%s%s", pr.Name, pr.Config.ProgramExt))
		return filepath.Join(progParts...)
	}

	if len(pr.allVersions) == 0 {
		return fmt.Errorf("missing plugin %s", pr.Source)
	}
	keySorter := &pluginKeySorter{pr.allVersions}
	sort.Sort(keySorter)
	for _, v := range pr.allVersions {
		if v.loadFunction != nil {
			if buf, err := v.loadFunction(context.Background()); err == nil {
				os.MkdirAll(filepath.Join(append(parts, v.Version, v.OsArch)...), 0776)
				if err = os.WriteFile(progName(v.Version), buf, 0776); err == nil {
					pr.Version = v.Version
					pr.PluginProgram = progName((pr.Version))
					return nil
				}
			}
		} else {
			pr.Version = v.Version
			pr.PluginProgram = progName((pr.Version))
			return nil
		}
	}

	return fmt.Errorf("missing plugin %s/%s/%s", pr.Hostname, pr.Namespace, pr.Name)
}

func (pr *PluginSource) evalLocalPluginProgram() error {
	localVersions, err := pr.ListLocalPluginVersions()
	if err != nil {
		return err
	}
	if pr.allVersions == nil {
		pr.allVersions = localVersions
	} else {
		pr.allVersions = append(pr.allVersions, localVersions...)
	}
	return nil
}

func (ps *PluginSource) evalRemotePluginProgram() error {
	bucket := aws.String("synwork-plugins")
	//creds := credentials.NewStaticCredentials("AKIAQ2XVJXL4VSXOTI6U", "TAe/Jnx/OOVG8ZnvXKc2obO7iVXVk3oy9lLDQNx3", "")
	creds := credentials.AnonymousCredentials
	cfg := &aws.Config{
		Region:      aws.String("eu-central-1"),
		Credentials: creds,
	}
	sess, _ := session.NewSession(cfg)

	downloader := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})
	// Write the contents of S3 Object to the file
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: bucket,
		Key:    aws.String(filepath.Join(ps.Hostname, ps.Namespace, ps.Name, "versions.xml")),
	})
	if err != nil {
		return fmt.Errorf("failed to download file, %v", err)
	}
	versions, err := ParseVersionFile(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("failed to download file, %v", err)
	}

	for _, v := range versions.Version {
		if parsedVersions, err := version.NewVersion(v.Id); err == nil {
			pr := &PluginReference{
				PluginKey: PluginKey{
					Hostname:      ps.Hostname,
					Namespace:     ps.Namespace,
					Name:          ps.Name,
					Ext:           ps.Ext,
					Version:       v.Id,
					OsArch:        ps.OsArch,
					ParsedVersion: parsedVersions,
				},
			}
			if ps.VersionConstraint.Check(pr.ParsedVersion) {
				file := fmt.Sprintf("synwork-processor-%s_%s_%s", pr.Name, pr.Version, pr.OsArch)
				pr.loadFunction = func(ctx context.Context) ([]byte, error) {
					buf := aws.NewWriteAtBuffer([]byte{})
					// Write the contents of S3 Object to the file
					_, err := downloader.Download(buf, &s3.GetObjectInput{
						Bucket: bucket,
						Key:    aws.String(filepath.Join(ps.Hostname, ps.Namespace, ps.Name, pr.Version, file)),
					})
					if err != nil {
						return nil, fmt.Errorf("failed to download file %s, %v", file, err)
					}

					return buf.Bytes(), nil
				}
				ps.allVersions = append(ps.allVersions, pr)
			}
		}

	}

	return nil
}
