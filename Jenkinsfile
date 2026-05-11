pipeline {
    agent any
    
    environment {
        NAMESPACE = 'app-namespace'
        MANIFESTS_DIR = 'k8s_manifests2'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Deploy') {
            agent {
                docker {
                    image 'bitnami/kubectl:latest'
                    args '-v /root/.kube:/home/user/.kube:ro -v ${WORKSPACE}:${WORKSPACE}'
                    alwaysPull true
                }
            }
            steps {
                script {
                    sh '''
                        kubectl config current-context
                        kubectl apply -f ${MANIFESTS_DIR}/namespace.yaml
                        kubectl apply -f ${MANIFESTS_DIR}/
                        kubectl wait --for=condition=available deployment --all -n ${NAMESPACE} --timeout=120s
                        kubectl get pods -n ${NAMESPACE} -o wide
                    '''
                }
            }
        }
    }
}
