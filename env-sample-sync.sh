#!/usr/bin/env bash

set -e

for i in "$@"; do
  case $i in
    -e=*|--env-file=*)
      ENV_FILE_NAME="${i#*=}"
      shift # past argument=value
      ;;
    -s=*|--sample-file=*)
      SAMPLE_FILE_NAME="${i#*=}"
      shift # past argument=value
      ;;
    -*|--*)
      echo "Unknown option $i"
      exit 1
      ;;
    *)
      ;;
  esac
done

function scrub_env_file() {
  local envFile=$1
  local sampleFile=$2
  local isComment='^[[:space:]]*#'
  local isBlank='^[[:space:]]*$'
  local sampleContent=""
  local headOfFile=1

  while IFS= read -r line; do
    # If the line contains an environment variable definition, scrub the variable value
    # and set a placeholder of the form: KEY=<KEY>
    if [[ ! $line =~ $isComment && ! $line =~ $isBlank ]]; then
      key=$(echo "$line" | cut -d '=' -f 1)
      sampleContent="${sampleContent}\n${key}=<${key}>"
    else
      sampleContent="${sampleContent}\n${line}"
    fi

    headOfFile=0
  done < <( cat "$envFile" )

  echo -e $sampleContent > $sampleFile
}

repoDir=$(git rev-parse --show-toplevel)
envFilePath="$repoDir/$ENV_FILE_NAME"
sampleFilePath="$repoDir/$SAMPLE_FILE_NAME"


if [[ -f $envFilePath ]]; then
  scrub_env_file $envFilePath $sampleFilePath
  git add $SAMPLE_FILE_NAME
else
  echo "No .env file found"
fi
