package kubernetes

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/charmbracelet/log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	containerVolumeMounts = []corev1.VolumeMount{
		{
			Name:      "docker-sock",
			MountPath: fmt.Sprintf("/var/run/user/%s/docker", uid),
			ReadOnly:  true,
		},
		{
			Name:      "home",
			MountPath: "/home",
		},
	}
)

var (
	initialInitContainerPodTemplate *template.Template
	initContainerPodTemplate        *template.Template
	containerTemplate               *template.Template
)

const (
	uid                           = "10000"
	initialInitContainerPodScript = `
		echo 'Creating user home';
		if [ ! -d /home/{{ .Username }} ]; then
			mkdir /home/{{ .Username }};
			chown -R {{ .UID }}:{{ .UID }} /home/{{ .Username }};
		fi;
		if [ ! -d /home/linuxbrew ]; then
		  echo 'Copying homebrew installation...';
		  mv /opt/linuxbrew /home/linuxbrew;
		  chown -R {{ .UID }}:{{ .UID }} /home/linuxbrew;
		fi;
	`
	initContainerPodScript = `
		if [ ! -d /home/{{ .Username }} ]; then
			mkdir /home/{{ .Username }};
			chown -R {{ .UID }}:{{ .UID }} /home/{{ .Username }};
		fi;
		if [ ! -d /home/linuxbrew ]; then
			apt-get update; apt-get install -y curl git;
			NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)";
			/home/linuxbrew/.linuxbrew/bin/brew install curl git man-db
			chown -R {{ .UID }}:{{ .UID }} /home/linuxbrew;
		fi;
	`
	containerScript = `
		echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"; export PATH=/home/linuxbrew/.linuxbrew/opt/man-db/libexec/bin:$PATH; export HOMEBREW_FORCE_BREWED_CURL=1; export HOMEBREW_CURL_PATH=/home/linuxbrew/.linuxbrew/bin/curl' > /etc/profile.d/99-linuxbrew.sh;
		echo 'export LANG=en_US.UTF-8' > /etc/profile.d/99-set-lang.sh;
		echo 'export DOCKER_HOST=unix:///var/run/user/{{ .UID }}/docker/docker.sock' > /etc/profile.d/99-set-docker-host.sh;
		groupadd -g 1000 docker;
		useradd -d /home/{{ .Username }} -M {{ .Username }} -u {{ .UID }} -s "$([ -f /home/{{ .Username }}/.boombox_shell ] && cat /home/{{ .Username }}/.boombox_shell || echo /bin/bash)" -G docker;
		touch /tmp/ready;
		tail -f /dev/null;
	`
)

func init() {
	var err error

	initialInitContainerPodTemplate, err = template.New("initialInitContainerPodTemplate").Parse(initialInitContainerPodScript)
	if err != nil {
		log.Fatal("Error initializing initialInitContainerPodTemplate", "error", err)
	}

	initContainerPodTemplate, err = template.New("initContainerPodTemplate").Parse(initContainerPodScript)
	if err != nil {
		log.Fatal("Error initializing initContainerPodTemplate", "error", err)
	}

	containerTemplate, err = template.New("containerTemplate").Parse(containerScript)
	if err != nil {
		log.Fatal("Error initializing containerTemplate", "error", err)
	}
}

func getInitialPodPayload(namespace, name, image string, pvc *corev1.PersistentVolumeClaim) *corev1.Pod {
	var tmpl bytes.Buffer
	if err := initialInitContainerPodTemplate.Execute(&tmpl, map[string]string{"Username": name, "UID": uid}); err != nil {
		log.Error("Error executing initial pod init container template", "error", err)
		return nil
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			InitContainers: []corev1.Container{
				{
					Name:            "init",
					Image:           fmt.Sprintf("ivan/boombox-init:%s", image),
					ImagePullPolicy: corev1.PullIfNotPresent,
					Args:            []string{"/bin/sh", "-c", tmpl.String()},
					VolumeMounts:    containerVolumeMounts,
				},
			},
			Containers: getContainersPayload(name, image),
			Volumes:    getVolumesPayload(pvc),
		},
	}

}

func getPodPayload(namespace, name, image string, pvc *corev1.PersistentVolumeClaim) *corev1.Pod {
	var tmpl bytes.Buffer
	if err := initContainerPodTemplate.Execute(&tmpl, map[string]string{"Username": name, "UID": uid}); err != nil {
		log.Error("Error executing pod init container template", "error", err)
		return nil
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			InitContainers: []corev1.Container{
				{
					Name:            "init",
					Image:           fmt.Sprintf("ivan/boombox-box:%s", image),
					ImagePullPolicy: corev1.PullAlways,
					Args:            []string{"/bin/sh", "-c", tmpl.String()},
					VolumeMounts:    containerVolumeMounts,
				},
			},
			Containers: getContainersPayload(name, image),
			Volumes:    getVolumesPayload(pvc),
		},
	}
}

func getContainersPayload(name, image string) []corev1.Container {
	var tmpl bytes.Buffer
	if err := containerTemplate.Execute(&tmpl, map[string]string{"Username": name, "UID": uid}); err != nil {
		log.Error("Error executing pod init container template", "error", err)
		return []corev1.Container{}
	}
	truePtr := true

	return []corev1.Container{
		{
			Name:         image,
			Image:        fmt.Sprintf("ivan/boombox-box:%s", image),
			Stdin:        true,
			TTY:          true,
			Args:         []string{"/bin/sh", "-c", tmpl.String()},
			VolumeMounts: containerVolumeMounts,
			ReadinessProbe: &corev1.Probe{
				TimeoutSeconds:   1,
				FailureThreshold: 60,
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{"cat", "/tmp/ready"},
					},
				},
			},
		}, {
			Name:  "dind",
			Image: "docker:dind-rootless",
			SecurityContext: &corev1.SecurityContext{
				Privileged: &truePtr,
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "docker-sock",
					MountPath: "/var/run/user/1000",
				},
			},
		},
	}
}

func getVolumesPayload(pvc *corev1.PersistentVolumeClaim) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "home",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc.Name,
				},
			},
		},
		{
			Name: "docker-sock",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		},
	}
}
