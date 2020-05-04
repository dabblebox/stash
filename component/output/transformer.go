package output

const (
	TypeTerraform         = "terraform"
	TypeECSTaskEnv        = "ecs-task-env"
	TypeECSTaskInjectJson = "ecs-task-inject-json"
	TypeECSTaskInjectEnv  = "ecs-task-inject-env"
	TypeJSONObject        = "json"
	TypeExport            = "terminal-export"
	TypeOriginal          = "original"
	TypeFile              = "file"
)

type ITransformer interface {
	Transform(data []byte) ([]byte, error)
}

// GetTransformer ...
func GetTransformer(output string, fileType string) (ITransformer, error) {

	switch output {
	case TypeJSONObject:
		return JSONTransformer{
			fileType: fileType,
		}, nil
	case TypeExport:
		return ExportTransformer{
			fileType: fileType,
		}, nil
	case TypeECSTaskEnv:
		return TaskDefEnvTransformer{
			fileType: fileType,
		}, nil
	}

	return PasshroughTransformer{}, nil
}
