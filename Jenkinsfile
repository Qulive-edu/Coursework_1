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
                    echo "Проверка подключения к кластеру"
                    sh 'kubectl cluster-info'
                    
                    echo "Применение манифестов"
                    sh "kubectl apply -f ${MANIFESTS_DIR}/"
                    
                    echo "Ожидание готовности деплойментов"
                    sh "kubectl wait --for=condition=available deployment --all -n ${NAMESPACE} --timeout=180s"
                    
                    echo "Итоговый статус подов"
                    sh "kubectl get pods -n ${NAMESPACE} -o wide"
                }
            }
        }
    }
}
