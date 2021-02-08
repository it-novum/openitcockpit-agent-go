pipeline {
    agent any
    environment {
        CIBUILD = "1"
        ADVINST = "\"C:\\Program Files (x86)\\Caphyon\\Advanced Installer\\bin\\x86\\advinst.exe\""
    }
    stages {
        /*
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
                        BINNAME = 'openitcockpit-agent.exe'
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
                        BINNAME = 'openitcockpit-agent'
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
        */
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
                        BINNAME = 'openitcockpit-agent.exe'
                    }
                    stages {
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
                        BINNAME = 'openitcockpit-agent'
                        CGO_ENABLED = '0'
                    }
                    stages {
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
                        BINNAME = 'openitcockpit-agent'
                    }
                    stages {
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
        stage('Package') {
            environment {
                VERSION = sh(
                    returnStdout: true,
                    script: 'cat VERSION'
                ).trim()
            }
            parallel {
                stage('Linux') {
                    agent {
                        dockerfile {
                            filename 'linux.Dockerfile'
                            dir 'build/docker'
                            label 'linux'
                            args "-u root --privileged"
                        }
                    }
                    environment {
                        GOOS = 'linux'
                        BINNAME = 'openitcockpit-agent'
                    }
                    stages {
                        stage('amd64') {
                            environment {
                                GOARCH = 'amd64'
                                ARCH = 'amd64'
                            }
                            steps {
                                package_linux()
                            }
                        }
                        stage('386') {
                            environment {
                                GOARCH = '386'
                                ARCH = 'i386'
                            }
                            steps {
                                package_linux()
                            }
                        }
                        stage('arm64') {
                            environment {
                                GOARCH = 'arm64'
                                ARCH = 'arm64'
                            }
                            steps {
                                package_linux()
                            }
                        }
                        stage('arm') {
                            environment {
                                GOARCH = 'arm'
                                ARCH = 'arm'
                            }
                            steps {
                                package_linux()
                            }
                        }
                    }
                }
                stage('windows') {
                    agent {
                        label 'windows'
                    }
                    environment {
                        GOOS = 'windows'
                        BINNAME = 'openitcockpit-agent.exe'
                    }
                    stages {
                        stage('amd64') {
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                package_windows()
                            }
                        }
                        stage('386') {
                            environment {
                                GOARCH = '386'
                            }
                            steps {
                                package_windows()
                            }
                        }
                    }
                }
            }
        }
    }
}

def cleanup_windows() {
    powershell '& git.exe clean -f -d -x'
}

def cleanup() {
    sh 'git clean -f -d -x'
}

def test_windows() {
    cleanup_windows()

    catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\cache C:\\gopath', returnStatus: true
        bat 'cd C:\\ & go.exe get -u github.com/t-yuki/gocover-cobertura'
        bat "go.exe test -coverprofile=cover.out -timeout=120s ./..."
        bat 'gocover-cobertura.exe < cover.out > coverage.xml'
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\gopath C:\\cache', returnStatus: true
    }
}

def test() {
    cleanup()

    catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
        sh 'cd / && go get -u github.com/t-yuki/gocover-cobertura'
        sh "go test -coverprofile=cover.out -timeout=120s ./..."
        sh 'gocover-cobertura < cover.out > coverage.xml'
    }
}


def build_windows_binary() {
    cleanup_windows()

    catchError(buildResult: null, stageResult: 'FAILURE') {
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\cache C:\\gopath', returnStatus: true
        bat "mkdir release\\$GOOS\\$GOARCH"
        bat "go.exe build -o release/$GOOS/$GOARCH/$BINNAME main.go"
        bat script: 'robocopy.exe /MIR /NFL /NDL /NJH /NJS /nc /ns /np C:\\gopath C:\\cache', returnStatus: true
    }
    archiveArtifacts artifacts: "release/$GOOS/$GOARCH/**", fingerprint: true
    stash name: "release-$GOOS-$GOARCH", includes: "release/$GOOS/$GOARCH/**"
}

def build_binary() {
    cleanup()

    catchError(buildResult: null, stageResult: 'FAILURE') {
        sh "mkdir -p release/$GOOS/$GOARCH"
        sh "go build -o release/$GOOS/$GOARCH/$BINNAME main.go"
    }
    archiveArtifacts artifacts: "release/$GOOS/$GOARCH/**", fingerprint: true
    stash name: "release-$GOOS-$GOARCH", includes: "release/$GOOS/$GOARCH/**"
}

def package_linux() {
    cleanup()

    unstash name: "release-$GOOS-$GOARCH"

    sh "mkdir -p package/usr/bin package/etc/openitcockpit-agent/ release/packages/$GOOS"
    sh 'cp example/config_example.cnf package/etc/openitcockpit-agent/config.cnf'
    sh 'cp example/customchecks_example.cnf package/etc/openitcockpit-agent/customchecks.cnf'
    sh "cp release/linux/$GOARCH/$BINNAME package/usr/bin/$BINNAME"
    sh "chmod +x package/usr/bin/$BINNAME"
    sh """cd release/packages/$GOOS &&
        fpm -s dir -t deb -C ../../../package --name openitcockpit-agent --vendor 'it-novum GmbH' \\
        --license 'Apache License Version 2.0' --config-files etc/openitcockpit-agent \\
        --architecture $ARCH --maintainer '<daniel.ziegler@it-novum.com>' \\
        --description 'openITCOCKPIT Monitoring Agent and remote plugin executor.' \\
        --url 'https://openitcockpit.io' --before-install ../../../build/package/preinst.sh \\
        --after-install ../../../build/package/postinst.sh --before-remove ../../../build/package/prerm.sh  \\
        --version '$VERSION'"""
    sh """cd release/packages/$GOOS &&
        fpm -s dir -t rpm -C ../../../package --name openitcockpit-agent --vendor 'it-novum GmbH' \\
        --license 'Apache License Version 2.0' --config-files etc/openitcockpit-agent \\
        --architecture $ARCH --maintainer '<daniel.ziegler@it-novum.com>' \\
        --description 'openITCOCKPIT Monitoring Agent and remote plugin executor.' \\
        --url 'https://openitcockpit.io' --before-install ../../../build/package/preinst.sh \\
        --after-install ../../../build/package/postinst.sh --before-remove ../../../build/package/prerm.sh  \\
        --version '$VERSION'"""
    sh """cd release/packages/$GOOS &&
        fpm -s dir -t pacman -C ../../../package --name openitcockpit-agent --vendor 'it-novum GmbH' \\
        --license 'Apache License Version 2.0' --config-files etc/openitcockpit-agent \\
        --architecture $ARCH --maintainer '<daniel.ziegler@it-novum.com>' \\
        --description 'openITCOCKPIT Monitoring Agent and remote plugin executor.' \\
        --url 'https://openitcockpit.io' --before-install ../../../build/package/preinst.sh \\
        --after-install ../../../build/package/postinst.sh --before-remove ../../../build/package/prerm.sh  \\
        --version '$VERSION'"""

    archiveArtifacts artifacts: 'release/packages/**', fingerprint: true
}

def package_windows() {
    cleanup_windows()

    unstash name: "release-$GOOS-$GOARCH"

    bat "$ADVINST /edit \"build\\msi\\openitcockpit-agent.aip\" \\SetVersion \"$VERSION\""
    bat "$ADVINST /build \"build\\msi\\openitcockpit-agent.aip\""
}
