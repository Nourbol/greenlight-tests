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
                    sh 'go build -o api ./cmd/api'
                }
            }
        }

        stage('Test') {
            steps {
                script {
                    sh 'go test ./cmd/api'
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
