pipeline {
    agent any
    
    stages {
        stage("Build linux/amd64") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'linux'
                GOARCH = 'amd64'
                BINNAME = 'agent'
            }
            steps {
                build_binary()
            }
        }
        stage("Build linux/386") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'linux'
                GOARCH = '386'
                BINNAME = 'agent'
            }
            steps {
                build_binary()
            }
        }
        stage("Build linux/arm") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'linux'
                GOARCH = 'arm'
                BINNAME = 'agent'
            }
            steps {
                build_binary()
            }
        }
        stage("Build linux/arm64") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'linux'
                GOARCH = 'arm64'
                BINNAME = 'agent'
            }
            steps {
                build_binary()
            }
        }

        stage("Build windows/amd64") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'windows'
                GOARCH = 'amd64'
                BINNAME = 'agent.exe'
            }
            steps {
                build_binary()
            }
        }
        stage("Build windows/386") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'windows'
                GOARCH = '386'
                BINNAME = 'agent.exe'
            }
            steps {
                build_binary()
            }
        }
        stage("Build windows/arm") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'windows'
                GOARCH = 'arm'
                BINNAME = 'agent'
            }
            steps {
                build_binary()
            }
        }
        stage("Build darwin/amd64") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'darwin'
                GOARCH = 'amd64'
                BINNAME = 'agent'
            }
            steps {
                build_binary()
            }
        }
        stage("Build darwin/386") {
            agent {
                docker { 
                    image 'golang:buster'
                    args "-u root --privileged -v agentgocache:/root/go"
                }
            }
            environment {
                GOOS = 'darwin'
                GOARCH = '386'
                BINNAME = 'agent'
            }
            steps {
                build_binary()
            }
        }
    }
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
    script {
        stash includes: 'release/**', name: 'release'
    }
}
