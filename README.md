# WebMon - Monitor your Web Pages with Go/GAE

Here is a simple application to monitor latency and or errors
of your web pages.

The app is written in Go for GAE.

> WARNING! Work in progress - currently just skeleton

## Setup

To **properly** checkout source you must obey following structure:
```bash
cd 
mkdir -p src/github.com/hpaluch/
cd src/github.com/hpaluch/
git clone https://github.com/hpaluch/webmon-go.git
```

> REMEMBER! You must have parent directory structure
> exactly set to `src/github.com/hpaluch/` otherwise
> all local go imports like:
> ```go
> import (
>  ...
>	"github.com/hpaluch/webmon-go/..."
>  ...
> )
> ```
> Would fail!!!
> Please see discussion
> at https://cloud.google.com/appengine/docs/flexible/go/using-go-libraries


Install required components:

* Tested OS: `Ubuntu 16.04.3 LTS`, `x86_64`

* Install python 2.7 (or later 2.x) using:

  ```bash
  sudo apt-get install python2.7
  ```

* Download current Google Cloud SDK (formerly GAE SDK) from:
  https://cloud.google.com/appengine/docs/standard/go/download
  in my case
  https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-171.0.0-linux-x86_64.tar.gz 

* Unpack your archive somewhere for example under `/opt`
  (you might need root permission):

```bash
sudo mkdir /opt/gae
sudo chown $USER /opt/gae
tar xzf google-cloud-sdk-171.0.0-linux-x86_64.tar.gz -C /opt/gae
```
* Add newly created `/opt/google-cloud-sdk/` to your `PATH`,
  for example add this to your `~/.bashrc`:

```bash
export PATH=/opt/gae/google-cloud-sdk/bin:$PATH
```

* and reload environment using:

```bash
source ~/.bashrc
```

* add Go GAE plugin to your Google Cloud SDK:

```bash
gcloud components install app-engine-go
```

Create new application in GAE Dashboard:

* Go to your GAE Dashboard using this link:
  https://console.cloud.google.com/projectselector/appengine/create?lang=go
* Click on `Create` button
* Fill in unique _Project name_ (in my case `hp-webmon`)
* click on `Create` button
* confirm `us-central` as region
* click on `Cancel Tutorial` if it bugs you.

## Developing app

* to run this app locally use:
```bash
./run_dev.sh
```
* and go to URL: http://localhost:8080/
* to view cute Admin interface (something like "Dashboard Lite")
  use: http://localhost:8000

## Deploying app

For the first time you must register your Google Account to deploy app:

* configure your project ID (in my case `hp-webmon`)
  in you shell set variable `WEBMON_APP_ID` to your
  _Project ID_ you created in your GAE Dashboard.
  For example I added to my `~/.bashrc`

  ```bash
export WEBMON_APP_ID=hp-webmon
  ```

  For the first time set app id manually:

  ```bash
gcloud config set project $WEBMON_APP_ID
  ```

* then create your App (for the 1st time only):

  ```bash
gcloud app create
  ```

* configure your Google Account for GAE:
```bash
gcloud config set account YOUR_GOOGLE_ACCOUNT
```
* login with your GAE account:
```bash
gcloud auth login
```
* new browser window should appear:
  * login or confirm selected account
  * allow required permissions for `Google Cloud SDK`
* you should see page with title "You are now authenticated with the Google Cloud SDK!"

And finally:
* to deploy app run script:
```bash
./deploy.sh
```

# Mis tips

How to view traces:

* Go to Dashboard of your GAE project
* click on "View Traces" on interested URL in list
* than click on point in trace graph
* you should now see detailed profile of your request


# Resources

I used many resources to write this program including
(but no guarantee to be comprehensive!):

Most of them come from my own App:
* ZoList written in Go for GAE:
  https://github.com/hpaluch/zolist-go 

