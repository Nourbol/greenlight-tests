pipeline {
    agent any

    triggers {
        githubPush()
    }

    tools {
        go 'Go'
    }

    stages {
        stage('Build') {
            steps {
                script {
                    sh "go mod vendor"
                    sh "go build -o api ./cmd/api"
                }
            }
        }

        stage('Test') {
            steps {
                script {
                    sh "go test ./cmd/api"
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
