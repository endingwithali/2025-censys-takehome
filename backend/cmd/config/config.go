package config

type DBConfig struct {
	Connection_String string
}

type HostFileConfig struct {
	MaxSize  int
	Location string
}

type ServerConfigurations struct {
	DBConfig       DBConfig
	HostFileConfig HostFileConfig
	Port           string
}

func Load() ServerConfigurations {
	db := DBConfig{
		Connection_String: "host=localhost user=backend password=backendpassword dbname=censys2025 port=5432 sslmode=disable TimeZone=UTC",
	}
	host := HostFileConfig{
		MaxSize:  (25 << 20),
		Location: "./backend/snapshots",
	}

	return ServerConfigurations{
		DBConfig:       db,
		HostFileConfig: host,
		Port:           ":8080",
	}
}
