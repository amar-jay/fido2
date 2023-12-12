run:
	echo "running both frontend and backend"; \
	cd ./backend && go run main.go & \
	pwd & \
	cd ./frontend && bun dev --host & \
	wait; \
	echo "terminated"

v:
	lt --port 5173 -s frontend-fido & \
	lt --port 8080 -s backend-fido & \
	wait; \
	echo "tunnelling"
