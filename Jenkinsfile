pipeline {
    agent any
    stages {
        stage('Test') {
            environment {
                CGO_ENABLED = '0'
            }
            parallel {
                stage('windows') {
                    agent {
                        docker { 
                            image 'golang:windowsservercore'
                            args '-v agentgocache:C:\\cache'
                            label 'windows'
                        }
                    }
                    environment {
                        GOOS = 'windows'
                        BINNAME = 'agent.exe'
                    }
                    stages {
                        stage('amd64') {
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                test_windows()
                            }
                        }
                        stage('386') {
                            environment {
                                GOARCH = '386'
                            }
                            steps {
                                test_windows()
                            }
                        }
                    }
                }
                stage('linux') {
                    agent {
                        docker { 
                            image 'golang:buster'
                            args "-u root --privileged -v agentgocache:/go"
                            label 'linux'
                        }
                    }
                    environment {
                        GOOS = 'linux'
                        BINNAME = 'agent'
                    }
                    stages {
                        stage('amd64') {
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                test()
                            }
                        }
                        stage('386') {
                            environment {
                                GOARCH = '386'
                            }
                            steps {
                                test()
                            }
                        }
                    }
                }
            }
        }
        stage('Build') {
            parallel {
                stage('windows') {
                    agent {
                        docker { 
                            image 'golang:windowsservercore'
                            args '-v agentgocache:C:\\cache'
                            label 'windows'
                        }
                    }
                    environment {
                        GOOS = 'windows'
                        BINNAME = 'agent.exe'
                    }
                    stages {
                        stage('cleanup') {
                            bat "if exist release\\$GOOS rmdir release\\$GOOS /q /s"
                        }
                        stage('amd64') {
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                build_windows_binary()
                            }
                        }
                        stage('386') {
                            environment {
                                GOARCH = '386'
                            }
                            steps {
                                build_windows_binary()
                            }
                        }
                    }
                }
                stage('linux') {
                    agent {
                        docker { 
                            image 'golang:buster'
                            args "-u root --privileged -v agentgocache:/go"
                            label 'linux'
                        }
                    }
                    environment {
                        GOOS = 'linux'
                        BINNAME = 'agent'
                        CGO_ENABLED = '0'
                    }
                    stages {
                        stage('cleanup') {
                            sh "rm -rf release/$GOOS"
                        }
                        stage('amd64') {
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                build_binary()
                            }
                        }
                        stage('386') {
                            environment {
                                GOARCH = '386'
                            }
                            steps {
                                build_binary()
                            }
                        }
                        stage('arm') {
                            environment {
                                GOARCH = 'arm'
                            }
                            steps {
                                build_binary()
                            }

                        }
                        stage('arm64') {
                            environment {
                                GOARCH = 'arm64'
                            }
                            steps {
                                build_binary()
                            }
                        }
                    }
                }
                stage('darwin') {
                    agent {
                        docker { 
                            image 'golang:buster'
                            args "-u root --privileged -v agentgocache:/go"
                            label 'linux'
                        }
                    }
                    environment {
                        GOOS = 'darwin'
                        BINNAME = 'agent'
                    }
                    stages {
                        stage('cleanup') {
                            sh "rm -rf release/$GOOS"
                        }
                        stage('amd64') {
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                build_binary()
                            }
                        }
                    }
                }
            }
        }
    }
}


def test_windows() {
    try {
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\cache C:\\gopath', returnStatus: true
        bat 'cd C:\\ & go.exe get -u github.com/t-yuki/gocover-cobertura'
        bat "go.exe test -coverprofile=cover.out -timeout=120s ./..."
        bat 'gocover-cobertura.exe < cover.out > coverage.xml'
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\gopath C:\\cache', returnStatus: true
    } catch (err) {
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
    }
}

def test() {
    try {
        sh 'cd / && go get -u github.com/t-yuki/gocover-cobertura'
        sh "go test -coverprofile=cover.out -timeout=120s ./..."
        sh 'gocover-cobertura < cover.out > coverage.xml'
    } catch (err) {
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
    }
}


def build_windows_binary() {
    try {
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\cache C:\\gopath', returnStatus: true
        bat "mkdir release\\$GOOS\\$GOARCH"
        bat "go.exe build -o release/$GOOS/$GOARCH/$BINNAME main.go"
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\gopath C:\\cache', returnStatus: true
    } catch (err) {
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
    }
    archiveArtifacts artifacts: 'release/**', fingerprint: true
}

def build_binary() {
    try {
        sh "mkdir -p release/$GOOS/$GOARCH"
        sh "go build -o release/$GOOS/$GOARCH/$BINNAME main.go"
    } catch (err) {
        echo "Caught: ${err}"
        currentBuild.result = 'FAILURE'
    }
    archiveArtifacts artifacts: 'release/**', fingerprint: true
}
