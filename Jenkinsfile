pipeline {
    agent any

    tools {
        go 'Go'
    }

    stages {
        stage('Print Environment') {
            steps {
                script {
                    sh 'env'
                }
            }
        }
        
        stage('Build') {
            steps {
                script {
                    echo "PATH: ${env.PATH}"
                    def goHome = tool 'Go'
                    echo "${goHome}"
                    sh "ls -l ${goHome}/bin"
                    sh "${goHome}/bin/go build -o api ./cmd/api"
                }
            }
        }

        stage('Test') {
            steps {
                script {
                    def goHome = tool 'Go'
                    sh "${goHome}/bin/go test ./cmd/api"
                }
            }
        }
    }

    post {
        success {
            echo 'Build and test successful! Deploying...'
        }

        failure {
            echo 'Build or test failed!'
        }
    }
}
