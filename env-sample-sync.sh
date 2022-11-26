#!/usr/bin/env bash

set -e

# A mapping of environment variables names to the the values used in `env.sample` as examples
declare -A examples

for i in "$@"; do
  case $i in
    -e=*|--env-file=*)
      ENV_FILE_NAME="${i#*=}"
      shift
      ;;
    -s=*|--sample-file=*)
      SAMPLE_FILE_NAME="${i#*=}"
      shift
      ;;
    -x=*|--example=*)
      varExample="${i#*=}"
      IFS='='; exampleTuple=($varExample); unset IFS;
      examples[${exampleTuple[0]}]=${exampleTuple[1]}
      shift
      ;;
    -*|--*)
      echo "Unknown option $i"
      exit 1
      ;;
    *)
      ;;
  esac
done

ENV_FILE_NAME=${ENV_FILE_NAME:-.env}
SAMPLE_FILE_NAME=${SAMPLE_FILE_NAME:-env.sample}

function scrub_env_file() {
  local envFile=$1
  local sampleFile=$2
  local isComment='^[[:space:]]*#'
  local isBlank='^[[:space:]]*$'
  local sampleContent=""

  while IFS= read -r line; do
    # If the line contains an environment variable definition, scrub the variable value
    # and set a placeholder of the form: ENV_VAR=<ENV_VAR>
    if [[ ! $line =~ $isComment && ! $line =~ $isBlank ]]; then
      envVar=$(echo "$line" | cut -d '=' -f 1)

      # If the user supplied an example for this environment variable, use their example
      # instead of the environemnt variable name
      if [[ ${examples[$envVar]+FOUND} == "FOUND" ]]; then
        exampleValue="${examples[$envVar]}"
      else
        exampleValue="<$envVar>"
      fi

      sampleContent="${sampleContent}${envVar}=$exampleValue\n"
    else
      sampleContent="${sampleContent}${line}\n"
    fi
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
  echo "No .env file found at: $envFilePath"
fi
