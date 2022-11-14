pipeline {
    environment {
      DOCKER = credentials('docker_hub')
    }
    agent any
        stages {
            stage('Build') {
                parallel {
                    stage('Express Image') {
                        steps {
                            sh 'docker build -f Dockerfile \
                            -t omollo/ekas-data-forwarding-prod:latest .'
                        }
                    }                    
                }
                post {
                    failure {
                        echo 'This build has failed. See logs for details.'
                    }
                }
            }
            stage('Test') {
                steps {
                    echo 'This is the Testing Stage'
                }
            }
            stage('DEPLOY') {
                when {
                    branch 'master'  //only run these steps on the master branch
                }
                steps {
                    // sh 'docker tag ekas-portal-api-dev:latest omollo/ekas-portal-api-prod:latest'
                    sh 'docker login -u "omollo" -p "safcom2012" docker.io'
                    sh 'docker push omollo/ekas-data-forwarding-prod:latest'
                }
            }
            stage('PUBLISH') {
                when {
                    branch 'master'  //only run these steps on the master branch
                }
                steps {
                    sh 'docker stack deploy -c docker-compose.yml ekas-data-forwarding-prod'
                }

            }

            // stage('REPORTS') {
            //     steps {
            //         junit 'reports.xml'
            //         archiveArtifacts(artifacts: 'reports.xml', allowEmptyArchive: true)
            //         // archiveArtifacts(artifacts: 'ekas-data-forwarding-prod-golden.tar.gz', allowEmptyArchive: true)
            //     }
            // }

            stage('CLEAN-UP') {
                steps {
                    // sh 'docker stop ekas-data-forwarding-dev'
                    sh 'docker service scale ekas-data-forwarding-prod_forwarding=0'
                    sh 'docker system prune -f'
                    sh 'docker service scale ekas-data-forwarding-prod_forwarding=2'
                    deleteDir()
                }
            }
        }
    }