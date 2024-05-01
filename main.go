package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/rox/generated/api/v1"
	storage "github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/roxctl/common"
	"github.com/stackrox/rox/roxctl/common/auth"
	roxctlIO "github.com/stackrox/rox/roxctl/common/io"
	"github.com/stackrox/rox/roxctl/common/logger"
	"github.com/stackrox/rox/roxctl/common/printer"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"io"
	"strings"
	"time"
)

func getImages(conn *grpc.ClientConn) ([]*storage.Image, error) {
	svc := v1.NewImageServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client, err := svc.ExportImages(ctx, &v1.ExportImageRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize stream client")
	}

	images := []*storage.Image{}
	for {
		image, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.Wrap(err, "stream broken by unexpected error")
		}

		images = append(images, image.Image)
	}

	return images, nil
}

func getDeployments(conn *grpc.ClientConn) ([]*storage.Deployment, error) {
	svc := v1.NewDeploymentServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	client, err := svc.ExportDeployments(ctx, &v1.ExportDeploymentRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize stream client")
	}

	deployments := []*storage.Deployment{}
	for {
		deployment, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, errors.Wrap(err, "stream broken by unexpected error")
		}

		deployments = append(deployments, deployment.Deployment)
	}

	return deployments, nil
}

func main() {
	defaultIO := roxctlIO.DefaultIO()
	conn, err := common.GetGRPCConnection(auth.TokenAuth(), logger.NewLogger(defaultIO, printer.DefaultColorPrinter()))
	if err != nil {
		panic(errors.Wrap(err, "could not establish gRPC connection to central"))
	}

	println("Fetching deployments")
	deployments, err := getDeployments(conn)
	if err != nil {
		panic(errors.Wrap(err, "could not get deployments"))
	}

	println("Fetching images")
	images, err := getImages(conn)
	if err != nil {
		panic(errors.Wrap(err, "could not get images"))
	}

	imageMap := map[string]*storage.Image{}

	for _, image := range images {
		imageMap[image.Name.FullName] = image
	}

	width, _, err := term.GetSize(0)
	if err != nil {
		panic(errors.Wrap(err, "could not get terminal size"))
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#84A59D"))).
		Width(width).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return lipgloss.NewStyle().Bold(true)
			case row%2 == 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("#EA9285"))
			default:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("#F5CAC3"))
			}
		}).
		Headers("CVE", "CVSS", "Cluster", "Namespace", "Image", "Component", "Fixable")

	for _, d := range deployments {
		for _, container := range d.Containers {
			imageName := container.Image.Name.FullName

			if strings.Contains(imageName, "openshift-release-dev") {
				continue
			}

			image, found := imageMap[imageName]
			if !found || image.Scan == nil {
				continue
			}

			if len(imageName) > 60 {
				imageName = imageName[:57] + "..."
			}

			for _, component := range image.Scan.Components {
				for _, vuln := range component.Vulns {
					score := "?"
					if vuln.CvssV3 != nil {
						score = fmt.Sprintf("v3: %.2f", vuln.CvssV3.Score)
					} else if vuln.CvssV2 != nil {
						score = fmt.Sprintf("v2: %.2f", vuln.CvssV2.Score)
					}

					t.Row(vuln.Cve, score, d.ClusterName, d.Namespace, imageName, component.Name, vuln.GetFixedBy())
				}
			}
		}
	}

	fmt.Println(t)
}
