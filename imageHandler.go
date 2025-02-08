package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Manifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

type ImageConfig struct {
	Config struct {
		Entrypoint []string `json:"Entrypoint"`
		Cmd        []string `json:"Cmd"`
	} `json:"Config"`
}

func ExtractImage(imageName string) ([]string, error) {
	tarPath := "./images/" + imageName + ".tar"

	manifestBytes, err := extractFile(tarPath, "manifest.json")
	if err != nil {
		panic(err)
	}
	var manifests []Manifest
	err = json.Unmarshal(manifestBytes, &manifests)
	if err != nil {
		panic(err)
	}
	if len(manifests) <= 0 {
		panic("No Manifests found")
	}
	manifest := manifests[0]

	configBytes, err := extractFile(tarPath, manifest.Config)
	if err != nil {
		panic(err)
	}
	var imageConfig ImageConfig
	err = json.Unmarshal(configBytes, &imageConfig)
	if err != nil {
		panic(err)
	}
	var cmd []string
	if len(imageConfig.Config.Entrypoint) > 0 {
		cmd = append(imageConfig.Config.Entrypoint, imageConfig.Config.Cmd...)
	} else {
		cmd = imageConfig.Config.Cmd
	}

	layerBytes, err := extractFile(tarPath, manifest.Layers[0])
	if err != nil {
		panic(err)
	}
	destDir := "./boxy-mcboxface/" + imageName
	untarLayer(layerBytes, destDir)

	return cmd, nil
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
		default:
			fmt.Printf("Skipping unsupported type: %c in file %s\n", header.Typeflag, header.Name)
		}
	}
	err = os.Chmod("./boxy-mcboxface/alpine/bin/busybox", 0755)
	if err != nil {
		panic(err)
	}
}

func extractFile(path string, target string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		if filepath.Clean(header.Name) == filepath.Clean(target) {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tarReader); err != nil {
				panic(err)
			}
			return buf.Bytes(), nil
		}
	}
	return nil, fmt.Errorf("file %s not found in tar archive", target)
}
