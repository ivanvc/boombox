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
	containerEnvVars = []corev1.EnvVar{
		{
			Name:  "HOMEBREW_FORCE_BREWED_CURL",
			Value: "1",
		},
		{
			Name:  "HOMEBREW_CURL_PATH",
			Value: "/home/linuxbrew/.linuxbrew/bin/curl",
		},
		{
			Name:  "LANG",
			Value: "en_US.UTF-8",
		},
	}
	containerVolumeMounts = []corev1.VolumeMount{
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
	initialInitContainerPodScript = `
		echo 'Creating user home';
		if [ ! -d /home/{{ .Username }} ]; then mkdir /home/{{ .Username }}; chown -R 10000:10000 /home/{{ .Username }}; fi;
		if [ ! -d /home/linuxbrew ]; then
		echo 'Copying homebrew installation...';
		mv /opt/linuxbrew /home/linuxbrew;
		chown -R 10000:10000 /home/linuxbrew;
		fi;
	`
	initContainerPodScript = `
		if [ ! -d /home/{{ .Username }} ]; then mkdir /home/{{ .Username }}; chown -R 10000:10000 /home/{{ .Username }}; fi;
		if [ ! -d /home/linuxbrew ]; then
			apt-get update; apt-get install -y curl git;
			NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)";
			/home/linuxbrew/.linuxbrew/bin/brew install curl git man-db
			chown -R 10000:10000 /home/linuxbrew;
		fi;
	`
	containerScript = `
		echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"; export PATH=/home/linuxbrew/.linuxbrew/opt/man-db/libexec/bin:$PATH' > /etc/profile.d/99-linuxbrew.sh;
		useradd -d /home/{{ .Username }} -M {{ .Username }} -u 10000 -s "$([ -f /home/{{ .Username }}/.boombox_shell ] && cat /home/{{ .Username }}/.boombox_shell || echo /bin/bash)";
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
	if err := initialInitContainerPodTemplate.Execute(&tmpl, map[string]string{"Username": name}); err != nil {
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
			Containers: []corev1.Container{getContainerPayload(name, image)},
			Volumes: []corev1.Volume{
				{
					Name: "home",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.ObjectMeta.Name,
						},
					},
				},
			},
		},
	}

}

func getPodPayload(namespace, name, image string, pvc *corev1.PersistentVolumeClaim) *corev1.Pod {
	var tmpl bytes.Buffer
	if err := initContainerPodTemplate.Execute(&tmpl, map[string]string{"Username": name}); err != nil {
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
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Args:            []string{"/bin/sh", "-c", tmpl.String()},
					VolumeMounts:    containerVolumeMounts,
				},
			},
			Containers: []corev1.Container{getContainerPayload(name, image)},
			Volumes: []corev1.Volume{
				{
					Name: "home",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.ObjectMeta.Name,
						},
					},
				},
			},
		},
	}
}

func getContainerPayload(name, image string) corev1.Container {
	var tmpl bytes.Buffer
	if err := containerTemplate.Execute(&tmpl, map[string]string{"Username": name}); err != nil {
		log.Error("Error executing pod init container template", "error", err)
		return corev1.Container{}
	}

	return corev1.Container{
		Name:         image,
		Image:        image,
		Stdin:        true,
		TTY:          true,
		Args:         []string{"/bin/sh", "-c", tmpl.String()},
		Env:          containerEnvVars,
		VolumeMounts: containerVolumeMounts,
	}
}
