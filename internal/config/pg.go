package config

import "fmt"

type Postgres struct {
	Host            string `mapstructure:"host" json:"host"`
	Port            string `mapstructure:"port" json:"port"`
	User            string `mapstructure:"user" json:"user"`
	Password        string `mapstructure:"password" json:"password"`
	DBName          string `mapstructure:"db_name" json:"db_name"`
	SSLMode         string `mapstructure:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time_sec" json:"conn_max_idle_time_sec"`
}

func (p *Postgres) GetDsn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port,
		p.User, p.Password,
		p.DBName, p.SSLMode)
}

func (p *Postgres) GetMaxOpenConns() int {
	if p.MaxOpenConns != 0 {
		return p.MaxOpenConns
	}
	panic("max open conns is not set")
}

func (p *Postgres) GetIdleConns() int {
	if p.MaxIdleConns != 0 {
		return p.MaxIdleConns
	}
	panic("max idle conns is not set")
}
func (p *Postgres) GetIdleTime() int {
	if p.ConnMaxIdleTime != 0 {
		return p.ConnMaxIdleTime
	}
	panic("max conn idle time is not set")
}
