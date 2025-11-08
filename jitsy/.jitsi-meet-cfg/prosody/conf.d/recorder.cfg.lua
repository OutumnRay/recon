-- Recorder configuration

-- Убедимся что модуль загружается для MUC компонента
local muc_modules = module:get_option_set("muc_modules", {})
muc_modules:add("muc_recorder_events")
module:set_option("muc_modules", muc_modules)

module:log("warn", "=== RECORDER CONFIG LOADED ===")
module:log("warn", "MUC modules: %s", tostring(muc_modules))