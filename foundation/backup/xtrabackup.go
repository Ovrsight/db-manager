package backup

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"log"
	"os/exec"
	"strings"
)

type XtraBackup struct {
}

type XtraBackupMock struct {
	mock.Mock
}

const ubuntu18Installer24Link = "https://downloads.percona.com/downloads/Percona-XtraBackup-2.4/Percona-XtraBackup-2.4.29/binary/debian/bionic/x86_64/percona-xtrabackup-24_2.4.29-1.bionic_amd64.deb"
const ubuntu20Installer24Link = "https://downloads.percona.com/downloads/Percona-XtraBackup-2.4/Percona-XtraBackup-2.4.29/binary/debian/focal/x86_64/percona-xtrabackup-24_2.4.29-1.focal_amd64.deb"
const ubuntu22Installer24Link = "https://downloads.percona.com/downloads/Percona-XtraBackup-2.4/Percona-XtraBackup-2.4.29/binary/debian/jammy/x86_64/percona-xtrabackup-24_2.4.29-1.jammy_amd64.deb"
const ubuntu20Installer8Link = "https://downloads.percona.com/downloads/Percona-XtraBackup-8.0/Percona-XtraBackup-8.0.35-30/binary/debian/focal/x86_64/percona-xtrabackup-80_8.0.35-30-1.focal_amd64.deb"
const ubuntu22Installer8Link = "https://downloads.percona.com/downloads/Percona-XtraBackup-8.0/Percona-XtraBackup-8.0.35-30/binary/debian/jammy/x86_64/percona-xtrabackup-80_8.0.35-30-1.jammy_amd64.deb"

func (xb *XtraBackup) Generate() ([]byte, error) {

	// detect that there is an installation of mysql
	program, err := exec.LookPath("mysqld")
	if err != nil {
		log.Fatal(err)
	}
	// detect the version of installed mysql
	cmd := exec.Command(fmt.Sprintf("%s", program), "--version")

	out, err := cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	var (
		beforeVersions string
		verKey         string
		afterVersions  string
		versionMajor   int
		versionMinor   int
		versionFix     int
	)

	_, err = fmt.Sscanf(string(out), "%s  %s %d.%d.%d%s", &beforeVersions, &verKey, &versionMajor, &versionMinor, &versionFix, &afterVersions)
	if err != nil {
		fmt.Println(string(out))
		log.Fatal(err)
	}

	// detect that there is an installed percona xtrabackup
	xtraBackupProgram, err := exec.LookPath("xtrabackup")
	if err != nil {

		lsbProgram, err := exec.LookPath("lsb_release")
		if err != nil {
			log.Fatal(err)
		}

		cmd := exec.Command(fmt.Sprintf("%s", lsbProgram), "-d")

		out, err = cmd.Output()
		if err != nil {
			log.Fatal(err, "here")
		}

		var (
			description    string
			distro         string
			releaseVersion string
			releaseName    string
		)

		_, err = fmt.Sscanf(string(out), "%s\t%s %s %s", &description, &distro, &releaseVersion, &releaseName)

		if strings.ToLower(distro) != "ubuntu" {
			log.Fatal(errors.New("only Ubuntu distro supported"))
		}

		err = xb.installXtraBackup(fmt.Sprintf("%d.%d", versionMajor, versionMinor), releaseVersion)
		if err != nil {
			log.Fatal(err)
		}
	}

	_ = xtraBackupProgram
	// TODO: work on creating anc compressing backup

	// create full backup
	// prepare backup
	// compress backup
	// read backup file data and return it for it to be uploaded

	return nil, nil
}

func (xb *XtraBackup) getXtraBackupInstallerUrlAndDependencies(ubuntuVersion, mysqlVersion string) (string, string, []string, error) {

	versions := strings.Split(ubuntuVersion, ".")

	minorVersion := strings.Join(versions[:2], ".")

	switch {
	case minorVersion == "18.04" && (mysqlVersion == "5.7" || mysqlVersion == "5.6"):
		return ubuntu18Installer24Link, strings.Split(ubuntu18Installer24Link, "/x86_64/")[1], []string{"libdbd-mysql-perl", "libcurl4-openssl-dev", "rsync", "libaio1", "libcurl4", "libev4"}, nil
	case minorVersion == "20.04" && (mysqlVersion == "5.7" || mysqlVersion == "5.6"):
		return ubuntu20Installer24Link, strings.Split(ubuntu20Installer24Link, "/x86_64/")[1], []string{"libdbd-mysql-perl", "libcurl4-openssl-dev", "rsync", "libaio1", "libcurl4", "libev4"}, nil
	case minorVersion == "22.04" && (mysqlVersion == "5.7" || mysqlVersion == "5.6"):
		return ubuntu22Installer24Link, strings.Split(ubuntu22Installer24Link, "/x86_64/")[1], []string{"libdbd-mysql-perl", "libcurl4-openssl-dev", "rsync", "libaio1", "libcurl4", "libev4"}, nil
	case minorVersion == "18.04" && mysqlVersion == "8.0":
		return "", "", nil, errors.New("installer not available for the selected ubuntu version")
	case minorVersion == "20.04" && mysqlVersion == "8.0":
		return ubuntu20Installer8Link, strings.Split(ubuntu20Installer8Link, "/x86_64/")[1], []string{"libdbd-mysql-perl", "libcurl4-openssl-dev", "rsync", "zstd", "libaio1", "libcurl4", "libev4"}, nil
	case minorVersion == "22.04" && mysqlVersion == "8.0":
		return ubuntu22Installer8Link, strings.Split(ubuntu22Installer8Link, "/x86_64/")[1], []string{"libdbd-mysql-perl", "libcurl4-openssl-dev", "rsync", "zstd", "libaio1", "libcurl4", "libev4"}, nil
	default:
		return "", "", nil, errors.New("installer not available for the selected ubuntu version")
	}
}

func (xb *XtraBackup) installXtraBackup(mysqlVersion, ubuntuVersion string) error {

	aptProgram, err := exec.LookPath("apt")
	if err != nil {
		log.Fatal(err)
	}

	// apt update
	cmd := exec.Command(fmt.Sprintf("%s", aptProgram), "update")

	out, err := cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	fmt.Println("Update result", string(out))

	// apt install wget
	cmd = exec.Command(fmt.Sprintf("%s", aptProgram), "install", "-y", "wget", "tar", "gzip")

	out, err = cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	wgetProgram, err := exec.LookPath("wget")
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command(fmt.Sprintf("%s", wgetProgram))

	installerUrl, installerName, dependencies, err := xb.getXtraBackupInstallerUrlAndDependencies(ubuntuVersion, mysqlVersion)
	if err != nil {
		log.Fatal(err)
	}

	switch mysqlVersion {
	case "5.6", "5.7", "8.0":
		cmd.Args = append(cmd.Args, installerUrl)
	default:
		return fmt.Errorf("mysql %s version is not yet supported", mysqlVersion)
	}

	out, err = cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	dependencies = append([]string{"install", "-y"}, dependencies...)

	cmd = exec.Command(fmt.Sprintf("%s", aptProgram), dependencies...)

	out, err = cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	dpkgProgram, err := exec.LookPath("dpkg")
	if err != nil {
		log.Fatal(err)
	}

	// dpkg -i percona-xtrabackup-80_8.0.35-30-1.jammy_amd64.deb
	cmd = exec.Command(fmt.Sprintf("%s", dpkgProgram), "-i", installerName)

	out, err = cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	// install tar & gz if they are not already installed
	cmd = exec.Command(fmt.Sprintf("%s", aptProgram), "tar", "gzip")

	out, err = cmd.Output()
	if err != nil {
		execErr := &exec.ExitError{}
		errors.As(err, &execErr)
		log.Fatalln(string(execErr.Stderr))
	}

	return nil
}

func (xbm *XtraBackupMock) Generate() ([]byte, error) {

	args := xbm.Called()
	return args.Get(0).([]byte), args.Error(1)
}
