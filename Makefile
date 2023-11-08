run:
	env GOOS=linux go build -o cli app/cli/*
	scp cli ubuntu@192.168.64.18:/home/ubuntu/cli
	scp .env ubuntu@192.168.64.18:/home/ubuntu/.env
	ssh ubuntu@192.168.64.18 './cli'