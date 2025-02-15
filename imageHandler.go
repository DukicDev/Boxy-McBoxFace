package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type FatManifests struct {
	Manifests []struct {
		Annotations struct {
			Arch string `json:"com.docker.official-images.bashbrew.arch"`
		} `json:"annotations"`
		Digest string `json:"digest"`
	} `json:"manifests"`
}

type Manifest struct {
	Config struct {
		Digest string `json:"digest"`
	} `json:"Config"`
	Layers []struct {
		Digest string `json:"digest"`
	} `json:"Layers"`
}

type ImageConfig struct {
	Config Config `json:"config"`
}

type Config struct {
	Entrypoint []string `json:"Entrypoint"`
	Cmd        []string `json:"Cmd"`
	WorkingDir string   `json:"WorkingDir"`
	Env        []string `json:"Env"`
}

func (c *Config) getEnvMap() map[string]string {
	envMap := make(map[string]string)
	for _, env := range c.Env {
		parts := strings.SplitN(env, "=", 2)
		envMap[parts[0]] = parts[1]
	}
	return envMap
}

type TokenResp struct {
	Token string `json:"token"`
}

var registryBaseUrl = "https://registry-1.docker.io/v2/library/"

func pullImage(imageName string, tag string) (Config, error) {
	bearerToken := getAuthToken(imageName)

	req, err := http.NewRequest("GET", registryBaseUrl+imageName+"/manifests/"+tag, nil)
	if err != nil {
		return Config{}, err
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return Config{}, err
	}
	defer resp.Body.Close()

	var fatManifests FatManifests
	if err := json.NewDecoder(resp.Body).Decode(&fatManifests); err != nil {
		return Config{}, err
	}
	arch := runtime.GOARCH
	var manifestSha string
	for _, manifest := range fatManifests.Manifests {
		if manifest.Annotations.Arch == arch {
			fmt.Printf("Found manifest for arch: %s with digest: %s\n", arch, manifest.Digest)
			manifestSha = manifest.Digest
			break
		}
	}
	req.URL.Path = registryBaseUrl + imageName + "/manifests/" + manifestSha
	resp, err = client.Do(req)
	if err != nil {
		return Config{}, err
	}
	defer resp.Body.Close()
	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return Config{}, err
	}

	req.URL.Path = registryBaseUrl + imageName + "/blobs/" + manifest.Config.Digest
	resp, err = client.Do(req)
	if err != nil {
		return Config{}, err
	}
	defer resp.Body.Close()
	var config ImageConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return Config{}, err
	}

	destDir := "/var/lib/boxy-mcboxface/containers/" + imageName
	err = os.MkdirAll(destDir, 0755)
	if err != nil {
		return Config{}, err
	}

	for _, layer := range manifest.Layers {
		var layerBytes []byte
		if layerBytes, err = os.ReadFile("/var/lib/boxy-mcboxface/layers/" + layer.Digest); err != nil {
			fmt.Printf("Pulling and extracting Layer %.16s...\n", strings.TrimPrefix(layer.Digest, "sha256:"))
			req.URL.Path = registryBaseUrl + imageName + "/blobs/" + layer.Digest
			resp, err = client.Do(req)
			if err != nil {
				return Config{}, err
			}
			defer resp.Body.Close()
			layerBytes, err = io.ReadAll(resp.Body)
			if err != nil {
				return Config{}, err
			}

			if err := os.MkdirAll("/var/lib/boxy-mcboxface/layers", 0700); err != nil {
				fmt.Printf("Could not create layers directory: Error: %s\n", err)
			}
			if err := os.WriteFile("/var/lib/boxy-mcboxface/layers/"+layer.Digest, layerBytes, 0700); err != nil {
				fmt.Printf("Could not cache layer %.16s\nError: %s\n", strings.TrimPrefix(layer.Digest, "sha256:"), err)
			}
		} else {
			fmt.Printf("Extracting cached Layer %.16s...\n", strings.TrimPrefix(layer.Digest, "sha256:"))
		}
		untarLayer(layerBytes, destDir)
	}

	return config.Config, nil
}

func getAuthToken(imageName string) string {
	resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/" + imageName + ":pull")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var tokenResp TokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		panic(err)
	}
	return tokenResp.Token
}

func untarLayer(layerBytes []byte, destDir string) {
	byteReader := bytes.NewReader(layerBytes)
	gzip, err := gzip.NewReader(byteReader)
	if err != nil {
		panic(err)
	}
	tarReader := tar.NewReader(gzip)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {

		case tar.TypeDir:
			err = os.MkdirAll(target, os.FileMode(header.Mode))
			if err != nil {
				panic(err)
			}

		case tar.TypeReg:
			err = os.MkdirAll(filepath.Dir(target), 0755)
			if err != nil {
				panic(err)
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				outFile.Close()
				panic(err)
			}
			outFile.Close()

		case tar.TypeSymlink:
			err = os.Symlink(header.Linkname, target)
			if err != nil {
				panic(err)
			}

		case tar.TypeLink:
			err = os.Link(filepath.Join(destDir, header.Linkname), target)
			if err != nil {
				panic(err)
			}
		default:
			fmt.Printf("Skipping unsupported type: %c in file %s\n", header.Typeflag, header.Name)
		}
	}

	resolveConf, err := os.OpenFile(filepath.Join(destDir, "etc/resolv.conf"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Couldnt create /etc/resolv.conf\n")
		return
	}
	defer resolveConf.Close()

	in, err := os.Open("/etc/resolv.conf")
	if err != nil {
		fmt.Printf("Couldnt read host resolv.conf\n")
		return
	}
	defer in.Close()

	_, err = io.Copy(resolveConf, in)
	if err != nil {
		fmt.Printf("Couldnt copy contents of %v to %v\n", in, resolveConf)
		return
	}

}
