pipeline {
    agent { label 'windows' }
    environment {
        NAMESPACE = 'app-namespace'
        MANIFESTS_DIR = 'k8s_manifests2'
    }

    stages {
        stage('Checkout Repository') {
            steps {
                checkout scm
                echo "Checked out: ${env.GIT_COMMIT?.take(7) ?: 'unknown'}"
            }
        }

        stage('Deploy to Minikube') {
            agent {
                docker {
                    image 'bitnami/kubectl:latest'
                    args '--network host'
                }
            }
            steps {
                script {
                    bat 'kubectl cluster-info || (echo "kubectl not configured for Minikube" && exit 1)'
                    
                    bat "kubectl apply -f ${MANIFESTS_DIR}/namespace.yaml"
                    
                    echo "Applying manifests..."
                    bat "kubectl apply -f ${MANIFESTS_DIR}/"
                    
                    echo "Waiting for deployments..."
                    bat "kubectl wait --for=condition=available deployment --all -n ${NAMESPACE} --timeout=60s"
                    
                    echo "Deployment status:"
                    bat "kubectl get pods -n ${NAMESPACE} -o wide"
                    
                }
            }

        }
    }
}