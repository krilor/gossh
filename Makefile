docker-up :
	cd testing; docker build --no-cache -t gossh_ubuntu -f Dockerfile .
	cd ..
	docker run -d --name gossh_ubuntu -p 2222:22 gossh_ubuntu

docker-down :
	( docker ps -a | grep gossh_ubuntu ) && docker kill gossh_ubuntu || true
	( docker ps -a | grep gossh_ubuntu ) && docker rm gossh_ubuntu || true
	ssh-keygen -f ~/.ssh/known_hosts -R "[localhost]:2222"

docker: docker-down docker-up
