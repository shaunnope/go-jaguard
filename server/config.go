// Defines the configuration structure for the Zouk cluster
package main

type Config struct {
	Servers []struct {
		Host string
		Port int
	}
}
