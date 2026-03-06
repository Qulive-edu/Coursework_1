pipeline {
    agent any

    environment {
        REGISTRY = "docker.io/qulive"
        CLIENT_IMAGE = "${REGISTRY}/stream-client"
        SERVER_IMAGE = "${REGISTRY}/stream-server"
    }

    stages {

        stage('Install Client Dependencies') {
            agent {
                docker {
                    image 'node:20'
                }
            }
            steps {
                dir('client') {
                    sh 'npm install'
                }
            }
        }

        stage('Build React Client') {
            agent {
                docker {
                    image 'node:20'
                }
            }
            steps {
                dir('client') {
                    sh 'npm run build'
                }
            }
        }

        stage('Build Go Server') {
            agent {
                docker {
                    image 'golang:1.22'
                }
            }
            steps {
                dir('stream-server') {
                    sh 'go build -o server'
                }
            }
        }

        stage('Build Docker Images') {
            steps {
                sh "docker build -t $CLIENT_IMAGE:latest ./client"
                sh "docker build -t $SERVER_IMAGE:latest ./stream-server"
            }
        }

        stage('Push Docker Images') {
            steps {
                withCredentials([usernamePassword(
                    credentialsId: 'dockerhub-credentials',
                    usernameVariable: 'DOCKER_USER',
                    passwordVariable: 'DOCKER_PASS'
                )]) {
                    sh 'echo $DOCKER_PASS | docker login -u $DOCKER_USER --password-stdin'
                    sh "docker push $CLIENT_IMAGE:latest"
                    sh "docker push $SERVER_IMAGE:latest"
                }
            }
        }

        stage('Deploy to Kubernetes') {
            steps {
                sh 'kubectl apply -f k8s_manifests2/'
            }
        }

        stage('Check Deployment') {
            steps {
                sh 'kubectl get pods'
                sh 'kubectl get services'
            }
        }
    }
}