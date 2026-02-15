// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package compose

import (
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/config"
)

// cloudRelayService returns the cloud relay service configuration
func cloudRelayService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "ghcr.io/trek-e/darkpipe/cloud-relay:${VERSION:-latest}",
		ContainerName: "darkpipe-relay",
		Restart:       "unless-stopped",
		Ports:         []string{"25:25"},
		Environment: map[string]string{
			"RELAY_HOSTNAME":              cfg.RelayHostname,
			"RELAY_DOMAIN":                cfg.MailDomain,
			"RELAY_LISTEN_ADDR":           "127.0.0.1:10025",
			"RELAY_TRANSPORT":             cfg.Transport,
			"RELAY_HOME_ADDR":             "10.8.0.2:25",
			"RELAY_MAX_MESSAGE_BYTES":     "52428800",
			"RELAY_EPHEMERAL_CHECK_INTERVAL": "60",
			"RELAY_QUEUE_ENABLED":         boolToString(cfg.QueueEnabled),
			"RELAY_QUEUE_KEY_PATH":        "/data/queue-keys/identity",
			"RELAY_QUEUE_SNAPSHOT_PATH":   "/data/queue-state/snapshot.json",
		},
		Volumes: []string{
			"postfix-queue:/var/spool/postfix",
			"certbot-etc:/etc/letsencrypt:ro",
			"queue-data:/data",
		},
		CapAdd:  []string{"NET_ADMIN"},
		Devices: []string{"/dev/net/tun"},
		Networks: []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "256M",
				},
				Reservations: &ComposeResourceLimits{
					Memory: "128M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "nc", "-z", "localhost", "25"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// stalwartService returns the Stalwart mail server service configuration
func stalwartService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "ghcr.io/trek-e/darkpipe/home-stalwart:${VERSION:-latest}",
		ContainerName: "stalwart",
		Restart:       "unless-stopped",
		Ports:         []string{"25:25", "587:587", "993:993", "8080:8080"},
		Environment: map[string]string{
			"MAIL_DOMAIN":        cfg.MailDomain,
			"MAIL_HOSTNAME":      cfg.RelayHostname,
			"ADMIN_EMAIL":        cfg.AdminEmail,
			"ADMIN_PASSWORD_FILE": "/run/secrets/admin_password",
		},
		Volumes: []string{
			"mail-data:/opt/stalwart-mail/data",
		},
		Secrets:  []string{"admin_password"},
		Networks: []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "512M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "nc", "-z", "localhost", "25"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// maddyService returns the Maddy mail server service configuration
func maddyService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "ghcr.io/trek-e/darkpipe/home-maddy:${VERSION:-latest}",
		ContainerName: "maddy",
		Restart:       "unless-stopped",
		Ports:         []string{"25:25", "587:587", "993:993"},
		Environment: map[string]string{
			"MAIL_DOMAIN":   cfg.MailDomain,
			"MAIL_HOSTNAME": cfg.RelayHostname,
		},
		Volumes: []string{
			"mail-data:/data",
		},
		Networks: []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "512M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "nc", "-z", "localhost", "25"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// postfixDovecotService returns the Postfix+Dovecot mail server service configuration
func postfixDovecotService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "ghcr.io/trek-e/darkpipe/home-postfix-dovecot:${VERSION:-latest}",
		ContainerName: "postfix-dovecot",
		Restart:       "unless-stopped",
		Ports:         []string{"25:25", "587:587", "993:993"},
		Environment: map[string]string{
			"MAIL_DOMAIN":   cfg.MailDomain,
			"MAIL_HOSTNAME": cfg.RelayHostname,
			"ADMIN_EMAIL":   cfg.AdminEmail,
		},
		Volumes: []string{
			"mail-data:/var/mail/vhosts",
			"mail-config:/etc/postfix-dovecot",
		},
		Networks: []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "512M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "nc", "-z", "localhost", "25"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// roundcubeService returns the Roundcube webmail service configuration
func roundcubeService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "roundcube/roundcubemail:1.6.13",
		ContainerName: "roundcube",
		Restart:       "unless-stopped",
		Ports:         []string{"8080:80"},
		Environment: map[string]string{
			"ROUNDCUBEMAIL_DB_TYPE":             "sqlite",
			"ROUNDCUBEMAIL_UPLOAD_MAX_FILESIZE": "25M",
		},
		Volumes: []string{
			"roundcube-data:/var/roundcube/db",
		},
		ExtraHosts: []string{"mail-server:host-gateway"},
		Networks:   []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "256M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "curl", "-f", "http://localhost/", "||", "exit", "1"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// snappymailService returns the SnappyMail webmail service configuration
func snappymailService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "djmaze/snappymail:2.38.2",
		ContainerName: "snappymail",
		Restart:       "unless-stopped",
		Ports:         []string{"8080:8888"},
		Environment: map[string]string{
			"UPLOAD_MAX_SIZE": "25M",
			"LOG_TO_STDOUT":   "true",
			"MEMORY_LIMIT":    "128M",
		},
		Volumes: []string{
			"snappymail-data:/var/lib/snappymail",
		},
		ExtraHosts: []string{"mail-server:host-gateway"},
		Networks:   []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "128M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8888/", "||", "exit", "1"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// radicaleService returns the Radicale CalDAV/CardDAV service configuration
func radicaleService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "tomsquest/docker-radicale:3.6.0",
		ContainerName: "radicale",
		Restart:       "unless-stopped",
		Ports:         []string{"5232:5232"},
		Volumes: []string{
			"radicale-data:/data/collections",
		},
		Networks: []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "128M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "curl", "-f", "http://localhost:5232/.web/", "||", "exit", "1"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// rspamdService returns the Rspamd spam filter service configuration
func rspamdService() *ComposeService {
	return &ComposeService{
		Image:         "rspamd/rspamd:latest",
		ContainerName: "rspamd",
		Restart:       "unless-stopped",
		Ports:         []string{"11334:11334"},
		Volumes: []string{
			"rspamd-data:/var/lib/rspamd",
		},
		DependsOn: []string{"redis"},
		Networks:  []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "256M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "nc", "-z", "localhost", "11332"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// redisService returns the Redis service configuration
func redisService() *ComposeService {
	return &ComposeService{
		Image:         "redis:alpine",
		ContainerName: "redis",
		Restart:       "unless-stopped",
		Volumes: []string{
			"redis-data:/data",
		},
		Networks: []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "64M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "redis-cli", "ping"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// caddyService returns the Caddy reverse proxy service configuration
func caddyService(cfg *config.Config) *ComposeService {
	return &ComposeService{
		Image:         "caddy:2-alpine",
		ContainerName: "caddy",
		Restart:       "unless-stopped",
		Ports:         []string{"80:80", "443:443", "443:443/udp"},
		Environment: map[string]string{
			"WEBMAIL_DOMAINS": "mail." + cfg.MailDomain,
		},
		Volumes: []string{
			"caddy-data:/data",
			"caddy-config:/config",
		},
		Networks: []string{"darkpipe"},
		Deploy: &ComposeDeploy{
			Resources: &ComposeResources{
				Limits: &ComposeResourceLimits{
					Memory: "128M",
				},
			},
		},
		HealthCheck: &ComposeHealthCheck{
			Test:     []string{"CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:2019/config/", "||", "exit", "1"},
			Interval: "30s",
			Timeout:  "10s",
			Retries:  3,
		},
	}
}

// boolToString converts a boolean to a string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
