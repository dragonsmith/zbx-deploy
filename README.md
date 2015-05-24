# Zabbix deploy helper

<a href="https://evilmartians.com/?utm_source=zbx-deploy">
<img src="https://evilmartians.com/badges/sponsored-by-evil-martians.svg" alt="Sponsored by Evil Martians" width="236" height="54"></a>

### Problem

Deployments of huge web apps make take forever — up 10 minutes and more, way more.
When deploying background task processing services like Sidekiq, you need to shutdown the service right before you start the deployment and boot it back when the deploy finishes.

If you use Zabbix to monitor your installation, it will probably fire an alarm if one of the services is down for a long time. That can trigger emails and SMS notifications. And false alarms are something that you need to eliminate as soon as possible.

### Solution

At [Evil Martians](https://evilmartians.com/?utm_source=zbx-deploy), we deploy a lot of Rails applications, small and huge, and we use Zabbix to monitor them. To avoid false alarms on deploy, we've created `zbx-deploy`, the Zabbix deploy helper.

Before the deploy, `zbx-deploy` will put all the hosts you deploy to into the [maintenance mode](https://www.zabbix.com/documentation/2.2/manual/maintenance). When the deploy finishes, `zbx-deploy` puts these hosts back into normal mode. By performing these actions, we can avoid false alarms triggered by long deploys.

`zbx-deploy` is a web service which provides two endpoints that need to be triggered on deploy:

```
POST /start/project-name # trigger it on deploy start
POST /finish/project-name # trigger it on deploy finish
```

## Deployment of the Zabbix helper application

To deploy the helper application itself, we've included a simple Makefile script:

```
ZBX_DEPLOY_TARGET=deploy@yourzabbix.com:/home/zbx-deploy/app/server make deploy
```

Where `ZBX_DEPLOY_TARGET` is the ENV variable with the host and directory to deploy to in scp-friendly format.
