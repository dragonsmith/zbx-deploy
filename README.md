# Zabbix deploy helper

### Problem
For a huge web app installations, deploy may take up to 10 minutes.
For some services like Sidekiq, you need to shutdown the service when the deploy starts and start it back when the deploy finishes.

When some app service is down for 10 minites, your Zabbix will probably fire an alarm, trigger emails ans SMS notifications.

### Solution

To avoid fake alarms on deploy, at [Evil Martians](http://evl.ms) we created `zbx-deploy`, the Zabbix deploy helper.

Before the deploy, `zbx-deploy` puts the hosts you deploy into [maintenance mode](https://www.zabbix.com/documentation/2.2/manual/maintenance). When the deploy is finished, `zbx-deploy` puts these hosts back into normal mode. By performing these actions, we can avoid fake alarms triggered by deploy operations.

`zbx-deploy` has two endpoints to be triggered on deploy:

```
POST /start/project-name # trigger it on deploy start
POST /finish/project-name # trigger it on deploy finish
```

## Deployment of this app

There is a simple Makefile script for deploy built-in to this app:

```
ZBX_DEPLOY_TARGET=deploy@yourzabbix.com:/home/zbx-deploy/app/server make release
```

`ZBX_DEPLOY_TARGET` is the ENV variable with the host and directory to deploy, set in scp-capable format.
