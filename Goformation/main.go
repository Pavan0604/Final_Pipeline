package main

import (
	"fmt"
	"github.com/awslabs/goformation/v6/cloudformation"
	"github.com/awslabs/goformation/v6/cloudformation/codebuild"
	"github.com/awslabs/goformation/v6/cloudformation/codepipeline"
	"github.com/awslabs/goformation/v6/cloudformation/ecr"
	"github.com/awslabs/goformation/v6/cloudformation/iam"
	"github.com/awslabs/goformation/v6/cloudformation/s3"
	"io"
	"os"
)

type AssumePolicyDocument struct {
	Statement []StatementEntry
}

type StatementEntry struct {
	Principal service
	Action    string
	Effect    string
}

type service struct {
	Service string
}

func main() {
	template := cloudformation.NewTemplate()
	template.AWSTemplateFormatVersion = "2010-09-09"
	template.Description = "The cloudformation template for Codepipeline"

	template.Parameters["Stage"] = cloudformation.Parameter{
		Type:    "String",
		Default: "dev",
	}
	template.Parameters["GithubUserName"] = cloudformation.Parameter{
		Type:    "String",
		Default: "Pavan0604",
	}
	template.Parameters["GithubRepo"] = cloudformation.Parameter{
		Type:    "String",
		Default: "Fargate-Codepipeline",
	}
	template.Parameters["GithubBranch"] = cloudformation.Parameter{
		Type:    "String",
		Default: "main",
	}
	template.Parameters["GithubOAuthToken"] = cloudformation.Parameter{
		Type:    "String",
		Default: "",
	}
	template.Parameters["ConatinerPort"] = cloudformation.Parameter{
		Type:    "Number",
		Default: 8000,
	}

	template.Resources["ECRRepository"] = &ecr.Repository{
		RepositoryName: cloudformation.String(cloudformation.Join("-", []string{cloudformation.Ref("Stage"), cloudformation.Ref("AWS::AccountId"), "ecr-repository"})),
	}

	template.Resources["S3Bucket"] = &s3.Bucket{
		BucketName: cloudformation.String(cloudformation.Join("-", []string{cloudformation.Ref("Stage"), cloudformation.Ref("AWS::AccountId"), "s3bucket"})),
	}

	template.Resources["CodePipeLineExecutionRole"] = &iam.Role{
		AssumeRolePolicyDocument: AssumePolicyDocument{
			Statement: []StatementEntry{
				{
					Effect: "Allow",
					Action: "sts:AssumeRole",
					Principal: service{
						Service: "codepipeline.amazonaws.com",
					},
				},
			},
		},
		ManagedPolicyArns: &[]string{
			"arn:aws:iam::aws:policy/AdministratorAccess",
		},
	}

	template.Resources["CodeBuildExecutionRole"] = &iam.Role{
		AssumeRolePolicyDocument: AssumePolicyDocument{
			Statement: []StatementEntry{
				{
					Effect: "Allow",
					Action: "sts:AssumeRole",
					Principal: service{
						Service: "codebuild.amazonaws.com",
					},
				},
			},
		},
		ManagedPolicyArns: &[]string{
			"arn:aws:iam::aws:policy/AdministratorAccess",
		},
	}

	template.Resources["CloudformationExecutionRole"] = &iam.Role{
		AssumeRolePolicyDocument: AssumePolicyDocument{
			Statement: []StatementEntry{
				{
					Effect: "Allow",
					Action: "sts:AssumeRole",
					Principal: service{
						Service: "cloudformation.amazonaws.com",
					},
				},
			},
		},
		ManagedPolicyArns: &[]string{
			"arn:aws:iam::aws:policy/AdministratorAccess",
		},
	}

	template.Resources["BuildProject"] = &codebuild.Project{
		Artifacts: &codebuild.Project_Artifacts{
			Type: "CODEPIPELINE",
		},
		Environment: &codebuild.Project_Environment{
			ComputeType:              "BUILD_GENERAL1_SMALL",
			PrivilegedMode:           cloudformation.Bool(true),
			Image:                    "aws/codebuild/standard:2.0",
			ImagePullCredentialsType: cloudformation.String("CODEBUILD"),
			Type:                     "LINUX_CONTAINER",
			EnvironmentVariables: &[]codebuild.Project_EnvironmentVariable{
				{
					Name:  "ECR_REPOSITORY_URI",
					Value: cloudformation.Join(".", []string{cloudformation.Ref("AWS::AccountId"), "dkr.ecr", cloudformation.Ref("AWS::Region"), cloudformation.Join("/", []string{"amazonaws.com", cloudformation.Ref("ECRRepository")})}),
				},
			},
		},
		Name:        cloudformation.String(cloudformation.Join("-", []string{cloudformation.Ref("Stage"), cloudformation.Ref("AWS::AccountId"), "BuildProject"})),
		ServiceRole: cloudformation.Ref("CodeBuildExecutionRole"),
		Source: &codebuild.Project_Source{
			Type:      "CODEPIPELINE",
			BuildSpec: cloudformation.String("buildspec.yml"),
		},
	}

	template.Resources["CodePipeLine"] = &codepipeline.Pipeline{
		AWSCloudFormationDependsOn: []string{"S3Bucket"},
		ArtifactStore: &codepipeline.Pipeline_ArtifactStore{
			Location: cloudformation.Join("-", []string{cloudformation.Ref("Stage"), cloudformation.Ref("AWS::AccountId"), "S3Bucket"}),
			Type:     "S3",
		},
		Name:                     cloudformation.String(cloudformation.Join("-", []string{cloudformation.Ref("Stage"), cloudformation.Ref("AWS::AccountId"), "CodePipeLine"})),
		RestartExecutionOnUpdate: cloudformation.Bool(false),
		RoleArn:                  cloudformation.GetAtt("CodePipeLineExecutionRole", "Arn"),

		Stages: []codepipeline.Pipeline_StageDeclaration{
			{
				Name: "Source",
				Actions: []codepipeline.Pipeline_ActionDeclaration{
					{
						Name: "Source",
						ActionTypeId: &codepipeline.Pipeline_ActionTypeId{
							Category: "Source",
							Owner:    "ThirdParty",
							Provider: "Github",
							Version:  "1",
						},
						RunOrder: cloudformation.Int(1),
						OutputArtifacts: &[]codepipeline.Pipeline_OutputArtifact{
							{
								Name: "source-output-artifacts",
							},
						},
					},
				},
			},
		},
	}

	y, err := template.YAML()
	if err != nil {
		fmt.Printf("Failed to generate YAML: %s\n", err)
	} else {
		content := string(y)
		filename, err := os.Create("./fargate.yaml")
		if err != nil {
			panic(err)
		}
		length, err := io.WriteString(filename, content)
		if err != nil {
			panic(err)
		}
		fmt.Println(length)
		defer filename.Close()
	}
}
