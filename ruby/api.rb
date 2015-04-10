require 'bundler/setup'
require "zabbixapi"

zbx = ZabbixApi.connect(
  url: ENV['ZBX_ENDPOINT'],
  user: ENV['ZBX_USERNAME'],
  password: ENV['ZBX_PASSWORD']
)

start_time = Time.now
period = 600
end_time = start_time + period

triggers = zbx.query(
  method: "maintenance.create",
  params: {
    "groupids"=>[12],
    "name"=>"demo",
    "maintenance_type"=>"0",
    "description"=>"created by zab.rb",
    "active_since"=>start_time.to_i,
    "active_till"=>end_time.to_i,
    "timeperiods"=> [{
        "timeperiod_type"=>"0",
        "start_date"=>start_time.to_i,
        "period"=>period.to_i
    }]
  }
)
puts triggers.inspect
