package awsprovider

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"sbl.systems/go/synwork/synwork/exts"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

const (
	baseUrl = "https://synwork-plugins.s3.eu-central-1.amazonaws.com/"
)

func init() {
	exts.RegisterProcessorProvider("awshttp", loadProcessorsFromAWShttp)
}

func loadProcessorsFromAWShttp(ctx context.Context, ps *exts.ProcessorKey) (exts.ProcessorSources, error) {
	path := filepath.Join(ps.Hostname, ps.Namespace, ps.Name, "versions.xml")
	resp, err := http.Get(baseUrl + path)
	if err != nil {
		return nil, err
	}
	versions, err := runtime.ParseVersionFile(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions, %v", err)
	}
	procSources := exts.ProcessorSources{}
	for _, versionItem := range versions.Version {
		file := fmt.Sprintf("synwork-processor-%s_%s_%s.html", ps.Name, versionItem.Id, ps.OsArch)
		file = filepath.Join(ps.Hostname, ps.Namespace, ps.Name, versionItem.Id, file)
		procSources = append(procSources, exts.ProcessorSource{
			Version: version.Must(version.NewVersion(versionItem.Id)),
			ProcessorProgram: func(ctx context.Context) ([]byte, error) {
				url := baseUrl + file
				resp, err := http.Get(url)
				if err != nil {
					return nil, err
				}
				return ReadHtmlEmbbededFile(ctx, resp.Body)
			},
		})
	}
	return procSources, nil
}

// func loadProcessorsFromAWSs3(ctx context.Context, ps *runtime.PluginKey) (exts.ProcessorSources, error) {
// 	bucket := aws.String("synwork-plugins")
// 	//creds := credentials.NewStaticCredentials("AKIAQ2XVJXL4VSXOTI6U", "TAe/Jnx/OOVG8ZnvXKc2obO7iVXVk3oy9lLDQNx3", "")
// 	creds := credentials.AnonymousCredentials
// 	cfg := &aws.Config{
// 		Region:      aws.String("eu-central-1"),
// 		Credentials: creds,
// 	}
// 	sess, _ := session.NewSession(cfg)

// 	downloader := s3manager.NewDownloader(sess)

// 	buf := aws.NewWriteAtBuffer([]byte{})
// 	// Write the contents of S3 Object to the file
// 	_, err := downloader.Download(buf, &s3.GetObjectInput{
// 		Bucket: bucket,
// 		Key:    aws.String(filepath.Join(ps.Hostname, ps.Namespace, ps.Name, "versions.xml")),
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to download file, %v", err)
// 	}
// 	versions, err := ParseVersionFile(bytes.NewReader(buf.Bytes()))
// 	if err != nil {
// 		return fmt.Errorf("failed to download file, %v", err)
// 	}

// 	for _, v := range versions.Version {
// 		if parsedVersions, err := version.NewVersion(v.Id); err == nil {
// 			pr := &PluginReference{
// 				PluginKey: PluginKey{
// 					Hostname:      ps.Hostname,
// 					Namespace:     ps.Namespace,
// 					Name:          ps.Name,
// 					Ext:           ps.Ext,
// 					Version:       v.Id,
// 					OsArch:        ps.OsArch,
// 					ParsedVersion: parsedVersions,
// 				},
// 			}
// 			if ps.VersionConstraint.Check(pr.ParsedVersion) {
// 				file := fmt.Sprintf("synwork-processor-%s_%s_%s", pr.Name, pr.Version, pr.OsArch)
// 				pr.loadFunction = func(ctx context.Context) ([]byte, error) {
// 					buf := aws.NewWriteAtBuffer([]byte{})
// 					// Write the contents of S3 Object to the file
// 					_, err := downloader.Download(buf, &s3.GetObjectInput{
// 						Bucket: bucket,
// 						Key:    aws.String(filepath.Join(ps.Hostname, ps.Namespace, ps.Name, pr.Version, file)),
// 					})
// 					if err != nil {
// 						return nil, fmt.Errorf("failed to download file %s, %v", file, err)
// 					}

// 					return buf.Bytes(), nil
// 				}
// 				ps.allVersions = append(ps.allVersions, pr)
// 			}
// 		}

// 	}
// 	return nil
// }
