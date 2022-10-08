package systemd

config_file: "/etc/my-service/config.yaml"

unit_file: '''
[Unit]
Description=My example service

[Service]
Environment=MY_CONFIG=\(config_file)
ExecStart=/usr/bin/my-service

[Install]
WantedBy=multi-user.target
'''
