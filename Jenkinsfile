pipeline {
    agent any
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
            steps {
                script {
                    sh 'kubectl cluster-info || (echo "kubectl not configured for Minikube" && exit 1)'
                    
                    sh "kubectl apply -f ${MANIFESTS_DIR}/namespace.yaml"
                    
                    echo "Applying manifests..."
                    sh "kubectl apply -f ${MANIFESTS_DIR}/"
                    
                    echo "Waiting for deployments..."
                    sh "kubectl wait --for=condition=available deployment --all -n ${NAMESPACE} --timeout=60s"
                    
                    echo "Deployment status:"
                    sh "kubectl get pods -n ${NAMESPACE} -o wide"
                    
                }
            }

        }
    }
}