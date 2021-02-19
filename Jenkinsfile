pipeline {
    agent any
    environment {
        CIBUILD = "1"
        ADVINST = "\"C:\\Program Files (x86)\\Caphyon\\Advanced Installer\\bin\\x86\\advinst.exe\""
    }
    stages {
        stage('Test') {
            environment {
                CGO_ENABLED = '0'
            }
            parallel {
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
                stage('darwin') {
                    agent {
                        label 'macos'
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
                        label 'macos'
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
                stage('darwin') {
                    agent {
                        label 'macos'
                    }
                    environment {
                        GOOS = 'darwin'
                        BINNAME = 'openitcockpit-agent'
                    }
                    stages {
                        stage('amd64') {
                            environment {
                                ARCH = 'amd64'
                                GOARCH = 'amd64'
                            }
                            steps {
                                package_darwin()
                            }
                        }
                    }
                }
            }
        }
    }
}

def cleanup_windows() {
    bat 'git.exe clean -f -d -x'
}

def cleanup() {
    sh 'git clean -f -d -x'
}

def test_windows() {
    timeout(time: 10, unit: 'MINUTES') {
        cleanup_windows()

        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
            bat 'cd C:\\ & go.exe get -u github.com/t-yuki/gocover-cobertura'
            bat "go.exe test -coverprofile=cover.out -timeout=120s ./..."
            bat 'gocover-cobertura.exe < cover.out > coverage.xml'
        }
    }
}

def test() {
    timeout(time: 10, unit: 'MINUTES') {
        cleanup()

        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
            sh 'cd / && go get -u github.com/t-yuki/gocover-cobertura'
            sh "go test -coverprofile=cover.out -timeout=120s ./..."
            sh 'gocover-cobertura < cover.out > coverage.xml'
        }
    }
}


def build_windows_binary() {
    timeout(time: 5, unit: 'MINUTES') {
        cleanup_windows()

        catchError(buildResult: null, stageResult: 'FAILURE') {
            bat "mkdir release\\$GOOS\\$GOARCH"
            bat "go.exe build -o release/$GOOS/$GOARCH/$BINNAME main.go"
        }
        archiveArtifacts artifacts: "release/$GOOS/$GOARCH/**", fingerprint: true
        stash name: "release-$GOOS-$GOARCH", includes: "release/$GOOS/$GOARCH/**"
    }
}

def build_binary() {
    timeout(time: 5, unit: 'MINUTES') {
        cleanup()

        catchError(buildResult: null, stageResult: 'FAILURE') {
            sh "mkdir -p release/$GOOS/$GOARCH"
            sh "go build -o release/$GOOS/$GOARCH/$BINNAME main.go"
        }
        archiveArtifacts artifacts: "release/$GOOS/$GOARCH/**", fingerprint: true
        stash name: "release-$GOOS-$GOARCH", includes: "release/$GOOS/$GOARCH/**"
    }
}

def package_linux() {
    timeout(time: 5, unit: 'MINUTES') {
        cleanup()

        unstash name: "release-$GOOS-$GOARCH"

        sh "mkdir -p package/usr/bin package/etc/openitcockpit-agent/ release/packages/$GOOS"
        sh 'cp example/config_example.ini package/etc/openitcockpit-agent/config.ini'
        sh 'cp example/customchecks_example.ini package/etc/openitcockpit-agent/customchecks.ini'
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
}

def package_windows() {
    timeout(time: 5, unit: 'MINUTES') {
        cleanup_windows()

        unstash name: "release-$GOOS-$GOARCH"

        // Convert Linux new lines to Windows new lines for older Windows Server systems
        bat 'move example\\config_example.ini example\\config_example_linux.ini'
        bat 'TYPE example\\config_example_linux.ini | MORE /P > example\\config_example.ini'

        bat 'move example\\customchecks_example.ini example\\customchecks_example_linux.ini'
        bat 'TYPE example\\customchecks_example_linux.ini | MORE /P > example\\customchecks_example.ini'

        powershell "& $ADVINST /loadpathvars \"build\\msi\\PathVariables_Jenkins.apf\""
        powershell "& $ADVINST /edit \"build\\msi\\openitcockpit-agent-${GOARCH}.aip\" \\SetVersion \"$VERSION\""
        powershell "& $ADVINST /build \"build\\msi\\openitcockpit-agent-${GOARCH}.aip\""
        archiveArtifacts artifacts: 'release/packages/**', fingerprint: true
    }
}

def package_darwin() {
    timeout(time: 5, unit: 'MINUTES') {
        cleanup()

        unstash name: "release-$GOOS-$GOARCH"

        sh "mkdir -p package/Applications/openitcockpit-agent package_osx_uninstaller release/packages/$GOOS"
        sh "cp release/$GOOS/$GOARCH/$BINNAME package/Applications/openitcockpit-agent/"
        sh "cp example/config_example.ini package/Applications/openitcockpit-agent/config.ini"
        sh "cp example/customchecks_example.ini package/Applications/openitcockpit-agent/customchecks.ini"
        sh "cp build/package/com.it-novum.openitcockpit.agent.plist package/Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist"
        sh """cd release/packages/$GOOS &&
            fpm -s dir -t osxpkg -C ../../../package --name openitcockpit-agent --vendor 'it-novum GmbH' \\
            --license "Apache License Version 2.0" --config-files Applications/openitcockpit-agent \\
            --architecture $ARCH --maintainer "<daniel.ziegler@it-novum.com>" \\
            --description "openITCOCKPIT Monitoring Agent and remote plugin executor." \\
            --url "https://openitcockpit.io" --before-install ../../../build/package/preinst.sh \\
            --after-install ../../../build/package/postinst.sh --version '$VERSION' &&
            mv openitcockpit-agent-${VERSION}.pkg openitcockpit-agent-${VERSION}-darwin-amd64.pkg"""
        sh """cd release/packages/$GOOS &&
            fpm -s dir -t osxpkg -C ../../../package_osx_uninstaller --name openitcockpit-agent-uninstaller --vendor "it-novum GmbH" \\
            --license "Apache License Version 2.0" --config-files Applications/openitcockpit-agent \\
            --architecture $ARCH --maintainer "<daniel.ziegler@it-novum.com>" \\
            --description "openITCOCKPIT Monitoring Agent and remote plugin executor." --url "https://openitcockpit.io" \\
            --before-install ../../../build/package/prerm.sh --version '$VERSION' --osxpkg-payload-free &&
            mv openitcockpit-agent-uninstaller-${VERSION}.pkg openitcockpit-agent-uninstaller-${VERSION}-darwin-amd64.pkg"""
        archiveArtifacts artifacts: 'release/packages/**', fingerprint: true
    }
}
