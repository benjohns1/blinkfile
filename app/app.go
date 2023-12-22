package app

type (
	Config struct {
		AdminCredentials Credentials
	}

	App struct {
		cfg Config
	}
)

func New(cfg Config) (*App, error) {
	return &App{cfg}, nil
}
