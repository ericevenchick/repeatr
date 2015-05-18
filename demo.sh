#!/bin/bash
set -eo pipefail

if [ -f .gopath/bin/repeatr ]; then PATH=$PWD/.gopath/bin/:$PATH; fi
if [ "$1" != "-t" ]; then straight=true; fi; export straight;
demodir="demo";

cnbrown="$(echo -ne "\E[0;33m")" # prompt
clblack="$(echo -ne "\E[1;30m")" # aside
clbrown="$(echo -ne "\E[1;33m")" # command
clblue="$(echo -ne "\E[1;34m")"  # section docs
cnone="$(echo -ne "\E[0m")"
awaitack() {
	[ "$straight" != true ] && return;
	echo -ne "${cnbrown}waiting for ye to hit enter, afore the voyage heave up anchor and make headway${cnone}"
	read -es && echo -ne "\E[F\E[2K\r"
}
tellRunning() {
	echo -e "${clblack}# running \`${clbrown}$@${clblack}\` >>>${cnone}"
}

rm -rf "$demodir"
mkdir -p "$demodir" && cd "$demodir" && demodir="$(pwd)"
echo "$demodir"




echo -e "${clblue}# Repeatr says hello!${cnone}"
echo -e "${clblue}#  Without a command, it provides help.${cnone}"
(
	tellRunning "repeatr"
	repeatr
)
echo -e "${clblue} ----------------------------${cnone}\n\n"
awaitack


echo -e "${clblue}# To suck in data, use the scan command:${cnone}"
echo
(
	tellRunning "repeatr scan --help"
	repeatr scan --help
	tellRunning "repeatr scan --kind=tar"
	repeatr scan --kind=tar
)
echo
echo -e "${clblue}#  This determines the data identity,${cnone}"
echo -e "${clblue}#   Uploads it to a warehouse,${cnone}"
echo -e "${clblue}#    And outputs the config to request it again later.${cnone}"
echo -e "${clblue} ----------------------------${cnone}\n\n"
awaitack




echo -e "${clblue}# The \`repeatr run\` command takes a job description and executes it.${cnone}"
echo -e "${clblue}#  Stdout goes to your terminal; any 'output' specifications are saved/uploaded.${cnone}"
(
	tellRunning "repeatr run -i some-json-config-files.conf"
	repeatr run -i <(cat <<-EOF
	{
		"Inputs": [
			{
				"Type": "tar",
				"Location": "/",
				"Hash": "b6nXWuXamKB3TfjdzUSL82Gg1avuvTk0mWQP4wgegscZ_ZzG9GfHDwKXQ9BfCx6v",
				"URI": "assets/ubuntu.tar.gz"
			}
		],
		"Accents": {
			"Entrypoint": [ "echo", "Hello from repeatr!" ]
		},
		"Outputs": [
			{
				"Type": "tar",
				"Location": "/var/log",
				"URI": "basic.tar"
			}
		]
	}
EOF
	)
)
echo -e "${clblue} ----------------------------${cnone}\n\n"
awaitack


echo "${clblue}#  That's all!  Neat, eh?${cnone}"


