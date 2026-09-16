pipeline {
    agent any 

    environment {
        NAMESPACE = 'app-namespace'
        MANIFESTS_DIR = 'k8s_manifests2'
        KUBECONFIG = '/var/jenkins_home/.kube/config'
    }

    stages {
        stage('Checkout Repository') {
            steps {
                checkout scm
                echo "Checked out: ${env.GIT_COMMIT?.take(7) ?: 'unknown'}"
            }
        }

        stage('Deploy to Kubernetes') {
            steps {
                script {
                    def KUBECONFIG_PATH = "/var/jenkins_home/.kube/config"
                    def K8S_CONTEXT = "docker-desktop" 
                    
                    echo "=== Проверка подключения к кластеру ==="
                    sh "kubectl --kubeconfig=${KUBECONFIG_PATH} --context=${K8S_CONTEXT} cluster-info"
                    
                    echo "=== Применение манифестов ==="
                    sh "kubectl --kubeconfig=${KUBECONFIG_PATH} --context=${K8S_CONTEXT} apply -f ${MANIFESTS_DIR}/"
                    
                    echo "=== Ожидание готовности деплойментов ==="
                    sh "kubectl --kubeconfig=${KUBECONFIG_PATH} --context=${K8S_CONTEXT} wait --for=condition=available deployment --all -n ${NAMESPACE} --timeout=180s"
                    
                    echo "=== Итоговый статус подов ==="
                    sh "kubectl --kubeconfig=${KUBECONFIG_PATH} --context=${K8S_CONTEXT} get pods -n ${NAMESPACE} -o wide"
                }
            }
        }
    }
}
