// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/cicada/readyz": {
            "get": {
                "description": "Check Cicada is ready",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Admin] System management"
                ],
                "summary": "Check Ready",
                "responses": {
                    "200": {
                        "description": "Successfully get ready state.",
                        "schema": {
                            "$ref": "#/definitions/pkg_api_rest_controller.SimpleMsg"
                        }
                    },
                    "500": {
                        "description": "Failed to check ready state.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/task_component": {
            "get": {
                "description": "Get a list of task component.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Component]"
                ],
                "summary": "List TaskComponent",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Page of the task component list.",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Row of the task component list.",
                        "name": "row",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get a list of task component.",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent"
                            }
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get a list of task component.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Register the task component.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Component]"
                ],
                "summary": "Create TaskComponent",
                "parameters": [
                    {
                        "description": "task component of the node.",
                        "name": "TaskComponent",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully register the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to register the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/task_component/{id}": {
            "get": {
                "description": "Get the task component.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Component]"
                ],
                "summary": "Get TaskComponent",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id of the TaskComponent",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "put": {
                "description": "Update the task component.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Component]"
                ],
                "summary": "Update TaskComponent",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id of the TaskComponent",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "task component to modify.",
                        "name": "TaskComponent",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully update the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to update the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete the task component.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Task Component]"
                ],
                "summary": "Delete TaskComponent",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the task component.",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully delete the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to delete the task component",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow": {
            "get": {
                "description": "Get a workflow list.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "List Workflow",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Page of the connection information list.",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Row of the connection information list.",
                        "name": "row",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get a workflow list.",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                            }
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get a workflow list.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a workflow.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Create Workflow",
                "parameters": [
                    {
                        "description": "Workflow content",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully create the workflow.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.WorkflowTemplate"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to create DAG.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow/run/{id}": {
            "post": {
                "description": "Run the workflow.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Run Workflow",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workflow ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully run the workflow.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to run the Workflow",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow/{id}": {
            "get": {
                "description": "Get the workflow.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Get Workflow",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the workflow.",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get the workflow.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get the workflow.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "put": {
                "description": "Update the workflow content.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Update Workflow",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workflow ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Workflow to modify.",
                        "name": "Workflow",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully update the workflow",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to update the workflow",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete the workflow.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow]"
                ],
                "summary": "Delete Workflow",
                "parameters": [
                    {
                        "type": "string",
                        "description": "ID of the workflow.",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully delete the workflow",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to delete the workflow",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow_template": {
            "get": {
                "description": "Get a list of workflow template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow Template]"
                ],
                "summary": "List WorkflowTemplate",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Page of the workflow template list.",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Row of the workflow template list.",
                        "name": "row",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get a list of workflow template.",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.WorkflowTemplate"
                            }
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get a list of workflow template.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        },
        "/workflow_template/{id}": {
            "get": {
                "description": "Get the workflow template.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "[Workflow Template]"
                ],
                "summary": "Get WorkflowTemplate",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id of the WorkflowTemplate",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Successfully get the workflow template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.WorkflowTemplate"
                        }
                    },
                    "400": {
                        "description": "Sent bad request.",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    },
                    "500": {
                        "description": "Failed to get the workflow template",
                        "schema": {
                            "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_common.ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Data": {
            "type": "object",
            "required": [
                "default_args",
                "task_groups"
            ],
            "properties": {
                "default_args": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.DefaultArgs"
                },
                "description": {
                    "type": "string"
                },
                "task_groups": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskGroup"
                    }
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.DefaultArgs": {
            "type": "object",
            "required": [
                "owner",
                "start_date"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "email_on_failure": {
                    "type": "boolean"
                },
                "email_on_retry": {
                    "type": "boolean"
                },
                "owner": {
                    "type": "string"
                },
                "retries": {
                    "description": "default: 1",
                    "type": "integer"
                },
                "retry_delay_sec": {
                    "description": "default: 300",
                    "type": "integer"
                },
                "start_date": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Params": {
            "type": "object",
            "required": [
                "properties",
                "required"
            ],
            "properties": {
                "properties": {},
                "required": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.ParmaOption": {
            "type": "object",
            "required": [
                "operator_option_for_use_as_param",
                "operator_option_value_is_json",
                "params"
            ],
            "properties": {
                "operator_option_for_use_as_param": {
                    "type": "string"
                },
                "operator_option_value_is_json": {
                    "type": "boolean"
                },
                "params": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Params"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Task": {
            "type": "object",
            "required": [
                "operator",
                "operator_options",
                "task_component",
                "task_name"
            ],
            "properties": {
                "dependencies": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "operator": {
                    "type": "string"
                },
                "operator_options": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "required": [
                            "name",
                            "value"
                        ],
                        "properties": {
                            "name": {
                                "type": "string"
                            },
                            "value": {}
                        }
                    }
                },
                "task_component": {
                    "type": "string"
                },
                "task_name": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskComponent": {
            "type": "object",
            "required": [
                "data",
                "id"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "data": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskData"
                },
                "id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskData": {
            "type": "object",
            "required": [
                "operator",
                "operator_options",
                "param_option",
                "task_name"
            ],
            "properties": {
                "operator": {
                    "type": "string"
                },
                "operator_options": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "required": [
                            "name",
                            "value"
                        ],
                        "properties": {
                            "name": {
                                "type": "string"
                            },
                            "value": {}
                        }
                    }
                },
                "param_option": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.ParmaOption"
                },
                "task_name": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.TaskGroup": {
            "type": "object",
            "required": [
                "task_group_name",
                "tasks"
            ],
            "properties": {
                "description": {
                    "type": "string"
                },
                "task_group_name": {
                    "type": "string"
                },
                "tasks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Task"
                    }
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Workflow": {
            "type": "object",
            "required": [
                "data",
                "id"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "data": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Data"
                },
                "id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "github_com_cloud-barista_cm-cicada_pkg_api_rest_model.WorkflowTemplate": {
            "type": "object",
            "required": [
                "data",
                "id"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "data": {
                    "$ref": "#/definitions/github_com_cloud-barista_cm-cicada_pkg_api_rest_model.Data"
                },
                "id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "pkg_api_rest_controller.SimpleMsg": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
