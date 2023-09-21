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
                    environment {
                        GOOS = 'linux'
                        BINNAME = 'openitcockpit-agent'
                    }
                    stages {
                        stage('amd64') {
                            agent {
                                docker { 
                                    image 'golang:bookworm'
                                    args "-u root --privileged -v agentgocache:/go"
                                    label 'linux'
                                }
                            }
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                test()
                            }
                        }
                        stage('386') {
                            agent {
                                docker { 
                                    image 'golang:bookworm'
                                    args "-u root --privileged -v agentgocache:/go"
                                    label 'linux'
                                }
                            }
                            environment {
                                GOARCH = '386'
                            }
                            steps {
                                test()
                            }
                        }
                        stage('arm64') {
                            agent {
                                docker { 
                                    image 'golang:bookworm'
                                    args "-u root --privileged -v agentgocache:/go"
                                    label 'linux-arm64'
                                }
                            }
                            environment {
                                GOARCH = 'arm64'
                            }
                            steps {
                                test()
                            }
                        }
                    }
                }
                stage('darwin') {

                    environment {
                        GOOS = 'darwin'
                        BINNAME = 'openitcockpit-agent'
                        CGO_ENABLED = '1'
                    }
                    stages {
                        stage('amd64') {
                            agent {
                                label 'macos'
                            }
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                test()
                            }
                        }
                        stage('arm64') {
                            agent {
                                label 'macos-arm64'
                            }
                            environment {
                                GOARCH = 'arm64'
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
                            image 'golang:bookworm'
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
                    environment {
                        GOOS = 'darwin'
                        BINNAME = 'openitcockpit-agent'
                    }
                    stages {
                        stage('amd64') {
                            agent {
                                label 'macos'
                            }
                            environment {
                                GOARCH = 'amd64'
                            }
                            steps {
                                build_binary()
                            }
                        }
                        stage('arm64') {
                            agent {
                                label 'macos-arm64'
                            }
                            environment {
                                GOARCH = 'arm64'
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
                                DEBARCH = 'amd64'
                                RPMARCH = 'amd64'
                            }
                            steps {
                                package_linux()
                            }
                        }
                        stage('386') {
                            environment {
                                GOARCH = '386'
                                ARCH = 'i386'
                                DEBARCH = 'i386'
                                RPMARCH = 'i386'
                            }
                            steps {
                                package_linux()
                            }
                        }
                        stage('arm64') {
                            environment {
                                GOARCH = 'arm64'
                                ARCH = 'arm64'
                                DEBARCH = 'arm64'
                                RPMARCH = 'aarch64'
                            }
                            steps {
                                package_linux()
                            }
                        }
                        stage('arm') {
                            environment {
                                GOARCH = 'arm'
                                ARCH = 'arm'
                                DEBARCH = 'armhf'
                                RPMARCH = 'armv7hl'
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
                    environment {
                        GOOS = 'darwin'
                        BINNAME = 'openitcockpit-agent'
                    }
                    stages {
                        stage('amd64') {
                            agent {
                                label 'macos'
                            }
                            environment {
                                ARCH = 'amd64'
                                GOARCH = 'amd64'
                            }
                            steps {
                                package_darwin_amd64()
                            }
                        }
                        stage('arm64') {
                            agent {
                                label 'macos-arm64'
                            }
                            environment {
                                ARCH = 'arm64'
                                GOARCH = 'arm64'
                            }
                            steps {
                                package_darwin_arm64()
                            }
                        }
                    }
                }
            }
        }
        stage('Publish') {
            when {
                branch 'main'
            }
            environment {
                VERSION = sh(
                    returnStdout: true,
                    script: 'cat VERSION'
                ).trim()
            }
            agent {
                label 'linux'
            }
            steps {
                publish_packages()
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
            bat "go.exe test -timeout=120s ./..."
        }
    }
}

def test() {
    timeout(time: 10, unit: 'MINUTES') {
        cleanup()

        catchError(buildResult: 'SUCCESS', stageResult: 'FAILURE') {
            sh "go test -timeout=120s ./..."
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

        sh "mkdir -p package/usr/bin package/etc/openitcockpit-agent/init release/packages/$GOOS package/var/log/openitcockpit-agent"
        sh 'cp example/config_example.ini package/etc/openitcockpit-agent/config.ini'
        sh 'cp example/customchecks_example.ini package/etc/openitcockpit-agent/customchecks.ini'
        sh 'cp build/package/openitcockpit-agent.init package/etc/openitcockpit-agent/init/openitcockpit-agent.init'
        sh 'cp build/package/openitcockpit-agent.service package/etc/openitcockpit-agent/init/openitcockpit-agent.service'
        sh "cp release/linux/$GOARCH/$BINNAME package/usr/bin/$BINNAME"
        sh "chmod +x package/usr/bin/$BINNAME"
        sh "chmod +x package/etc/openitcockpit-agent/init/openitcockpit-agent.init"
        sh """cd release/packages/$GOOS &&
            fpm -s dir -t deb -C ../../../package --name openitcockpit-agent --vendor 'it-novum GmbH' \\
            --license 'Apache License Version 2.0' --config-files etc/openitcockpit-agent \\
            --architecture $DEBARCH --maintainer '<daniel.ziegler@it-novum.com>' \\
            --description 'openITCOCKPIT Monitoring Agent and remote plugin executor.' \\
            --url 'https://openitcockpit.io' --before-install ../../../build/package/preinst.sh \\
            --after-install ../../../build/package/postinst.sh --before-remove ../../../build/package/prerm.sh  \\
            --version '$VERSION'"""
        sh """cd release/packages/$GOOS &&
            fpm -s dir -t rpm -C ../../../package --name openitcockpit-agent --vendor 'it-novum GmbH' \\
            --license 'Apache License Version 2.0' --config-files etc/openitcockpit-agent \\
            --architecture $RPMARCH --maintainer '<daniel.ziegler@it-novum.com>' \\
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

def package_darwin_amd64() {
    timeout(time: 5, unit: 'MINUTES') {
        cleanup()

        unstash name: "release-$GOOS-$GOARCH"

        sh "mkdir -p package/Applications/openitcockpit-agent package_osx_uninstaller release/packages/$GOOS"
        sh "cp release/$GOOS/$GOARCH/$BINNAME package/Applications/openitcockpit-agent/"
        sh "cp example/config_example.ini package/Applications/openitcockpit-agent/config.ini"
        sh "cp example/customchecks_example.ini package/Applications/openitcockpit-agent/customchecks.ini"
        sh "cp build/package/com.it-novum.openitcockpit.agent.plist package/Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist"
        sh "chmod +x package/Applications/openitcockpit-agent/$BINNAME"
        
        sh """/usr/local/bin/packagesbuild --package-version "${VERSION}" --reference-folder . build/macos/openITCOCKPIT\\ Monitoring\\ Agent/openITCOCKPIT\\ Monitoring\\ Agent.pkgproj"""
        sh """mv -f build/macos/openITCOCKPIT\\ Monitoring\\ Agent/build/openitcockpit-agent-darwin-amd64.pkg release/packages/${GOOS}/openitcockpit-agent-${VERSION}-darwin-${GOARCH}.pkg"""

        sh """cd release/packages/$GOOS &&
            fpm -s dir -t osxpkg -C ../../../package_osx_uninstaller --name openitcockpit-agent-uninstaller --vendor "it-novum GmbH" \\
            --license "Apache License Version 2.0" --config-files Applications/openitcockpit-agent \\
            --maintainer "<daniel.ziegler@it-novum.com>" \\
            --description "Uninstaller of openITCOCKPIT Monitoring Agent and remote plugin executor." --url "https://openitcockpit.io" \\
            --before-install ../../../build/package/prerm.sh --version '$VERSION' --osxpkg-payload-free &&
            mv openitcockpit-agent-uninstaller-${VERSION}.pkg openitcockpit-agent-uninstaller-${VERSION}-darwin-all.pkg"""

        archiveArtifacts artifacts: 'release/packages/**', fingerprint: true

        /*sh """cd release/packages/$GOOS &&
            fpm -s dir -t osxpkg -C ../../../package --name openitcockpit-agent --vendor 'it-novum GmbH' \\
            --license "Apache License Version 2.0" --config-files Applications/openitcockpit-agent \\
            --architecture $ARCH --maintainer "<daniel.ziegler@it-novum.com>" \\
            --description "openITCOCKPIT Monitoring Agent and remote plugin executor." \\
            --url "https://openitcockpit.io" --before-install ../../../build/package/preinst.sh \\
            --after-install ../../../build/package/postinst.sh --version '$VERSION' &&
            mv openitcockpit-agent-${VERSION}.pkg openitcockpit-agent-${VERSION}-darwin-${GOARCH}.pkg"""
        sh """cd release/packages/$GOOS &&
            fpm -s dir -t osxpkg -C ../../../package_osx_uninstaller --name openitcockpit-agent-uninstaller --vendor "it-novum GmbH" \\
            --license "Apache License Version 2.0" --config-files Applications/openitcockpit-agent \\
            --architecture $ARCH --maintainer "<daniel.ziegler@it-novum.com>" \\
            --description "openITCOCKPIT Monitoring Agent and remote plugin executor." --url "https://openitcockpit.io" \\
            --before-install ../../../build/package/prerm.sh --version '$VERSION' --osxpkg-payload-free &&
            mv openitcockpit-agent-uninstaller-${VERSION}.pkg openitcockpit-agent-uninstaller-${VERSION}-darwin-${GOARCH}.pkg"""
        archiveArtifacts artifacts: 'release/packages/**', fingerprint: true*/
    }
}

def package_darwin_arm64() {
    timeout(time: 5, unit: 'MINUTES') {
        cleanup()

        unstash name: "release-$GOOS-$GOARCH"

        sh "mkdir -p package/Applications/openitcockpit-agent package_osx_uninstaller release/packages/$GOOS"
        sh "cp release/$GOOS/$GOARCH/$BINNAME package/Applications/openitcockpit-agent/"
        sh "cp example/config_example.ini package/Applications/openitcockpit-agent/config.ini"
        sh "cp example/customchecks_example.ini package/Applications/openitcockpit-agent/customchecks.ini"
        sh "cp build/package/com.it-novum.openitcockpit.agent.plist package/Applications/openitcockpit-agent/com.it-novum.openitcockpit.agent.plist"
        sh "chmod +x package/Applications/openitcockpit-agent/$BINNAME"
        
        sh """/usr/local/bin/packagesbuild --package-version "${VERSION}" --reference-folder . build/macos/openITCOCKPIT\\ Monitoring\\ Agent\\ arm64/openITCOCKPIT\\ Monitoring\\ Agent.pkgproj"""
        sh """mv -f build/macos/openITCOCKPIT\\ Monitoring\\ Agent\\ arm64/build/openitcockpit-agent-darwin-arm64.pkg release/packages/${GOOS}/openitcockpit-agent-${VERSION}-darwin-${GOARCH}.pkg"""

        archiveArtifacts artifacts: 'release/packages/**', fingerprint: true
    }
}

def publish_packages() {
    timeout(time: 5, unit: 'MINUTES') {

        dir('publish'){
        /* get all packages */
            unarchive mapping: ['release/packages/' : '.']

            sh """mkdir packages"""
            sh """mv -f release/packages/darwin/* packages/"""
            sh """mv -f release/packages/linux/* packages/"""
            sh """mv -f release/packages/windows/* packages/"""

            sh 'ssh -o StrictHostKeyChecking=no -i /var/lib/jenkins/.ssh/id_rsa oitc@srvitnweb05.static.itsm.love "mkdir -p /var/www/openitcockpit.io/files/openitcockpit-agent-3.x"'
            sh 'rsync -avz -e "ssh -o StrictHostKeyChecking=no -i /var/lib/jenkins/.ssh/id_rsa" --delete --progress packages/* oitc@srvitnweb05.static.itsm.love:/var/www/openitcockpit.io/files/openitcockpit-agent-3.x/'

            /* Remove old packages */
            sh '/var/lib/jenkins/openITCOCKPIT-build/aptly.sh repo remove openitcockpit-agent-stable openitcockpit-agent'
            sh '/var/lib/jenkins/openITCOCKPIT-build/aptly.sh publish drop deb filesystem:openitcockpit-agent:deb/stable'

            /* Add packages to own apt repository */
            sh '/var/lib/jenkins/openITCOCKPIT-build/aptly.sh repo add -force-replace openitcockpit-agent-stable packages/openitcockpit-agent_*.deb;'

            /* Publish apt repository */
            sh '/var/lib/jenkins/openITCOCKPIT-build/aptly.sh publish repo -distribution deb -architectures amd64,i386,arm64,armhf -passphrase-file /opt/repository/aptly/pw -batch openitcockpit-agent-stable filesystem:openitcockpit-agent:deb/stable'
            sh "rsync -rv --delete-after /opt/repository/aptly/openitcockpit-agent/deb/stable/ www-data@srvoitcapt02.ad.it-novum.com:/var/www/html/openitcockpit-agent/deb/stable/"

            /* sign rpm packages */
            sh """rpmsign --define "_gpg_name it-novum GmbH <support@itsm.it-novum.com>" --define "__gpg_sign_cmd %{__gpg} gpg --no-verbose --no-armor --batch --pinentry-mode loopback --passphrase-file /opt/repository/aptly/pw %{?_gpg_digest_algo:--digest-algo %{_gpg_digest_algo}} --no-secmem-warning -u '%{_gpg_name}' -sbo %{__signature_filename} %{__plaintext_filename}" --addsign packages/*.rpm"""

            /* create yum repository */
            sh """mkdir -p rpm/stable"""
            sh """cp packages/*.rpm rpm/stable/"""
            /*
            createrepo is the old Python createrepo which does not exists on Ubuntu Focal.
            Ubuntu Focal has no createrepo_c.
            For this reason we use a Docker image with a recent version of createrepo_c which works fine
            https://github.com/it-novum/createrepo_c-docker
            */
            /*sh """createrepo rpm/stable"""*/
            sh """docker run --rm --user 111:116 -v '${env.WORKSPACE}/publish/rpm/stable':/rpm openitcockpit/createrepo_c"""
            /*sh """chown jenkins:jenkins ${env.WORKSPACE} -R"""*/

            /* Publish yum repository */
            sh "rsync -rv --delete-after rpm/stable/ www-data@srvoitcapt02.ad.it-novum.com:/var/www/html/openitcockpit-agent/rpm/stable/"
        }

    }
}
