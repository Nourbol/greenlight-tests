resource "aws_security_group" "web_server_sg" {
  name        = "web-server-sg"
  description = "Allow HTTP and SSH traffic"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "web_server" {
  ami             = "ami-053b0d53c279acc90"
  instance_type   = "t2.micro"
  key_name        = "nginx-terraform-nba"
  security_groups = [aws_security_group.web_server_sg.name]
  user_data       = <<-EOF
              #!/bin/bash
              sudo apt-get update
              sudo apt-get install -y docker.io
              docker pull nginx
              sudo docker run -d -p 80:80 nginx
              EOF
  tags            = {
    Name = "WebServer"
  }
}