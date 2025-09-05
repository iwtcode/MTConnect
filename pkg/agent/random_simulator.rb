require 'socket'
require 'rexml/document'
require 'time'

AGENT_HOST = '127.0.0.1'
DEVICES_XML_PATH = 'D:/MTConnect/agent-2.5.0.11-win64/demo/agent/Devices.xml'

PREDEFINED_VALUES = {
  "AVAILABILITY" => ["AVAILABLE"],
  "EXECUTION" => ["READY", "ACTIVE", "INTERRUPTED", "STOPPED"],
  "CONTROLLER_MODE" => ["AUTOMATIC", "MANUAL", "MANUAL_DATA_INPUT", "SEMI_AUTOMATIC"],
  "FUNCTIONAL_MODE" => ["PRODUCTION", "SETUP", "MAINTENANCE", "TEARDOWN"],
  "EMERGENCY_STOP" => ["ARMED", "TRIGGERED"],
  "DOOR_STATE" => ["OPEN", "CLOSED", "UNLATCHED"],
  "CHUCK_STATE" => ["OPEN", "CLOSED", "UNLATCHED"],
  "AXIS_STATE" => ["STOPPED", "TRAVELING", "HOMING"],
  "ROTARY_MODE" => ["SPINDLE", "INDEX", "CONTOUR"],
  "USER" => ["OperatorA", "MaintenanceGuy", "SetupPerson"],
  "EQUIPMENT_MODE" => ["ON", "OFF"],
  "PROGRAM_EDIT" => ["ACTIVE", "READY", "NOT_READY"],
  "WAIT_STATE" => ["WAITING", "BLOCKED"],
  "CONTROLLER_MODE_OVERRIDE" => ["ON", "OFF"],
  "e:OUTPUT_SIGNAL" => ["ON", "OFF"]
}

def parse_devices_xml(file_path)
  devices = {}
  begin
    file = File.new(file_path)
    doc = REXML::Document.new(file)
  rescue => e
    puts "Ошибка чтения или парсинга файла #{file_path}: #{e.message}"
    exit
  end

  doc.elements.each("MTConnectDevices/Devices/Device") do |device|
    device_name = device.attributes['name']
    port = device_name == 'OKUMA' ? 7878 : 7879
    devices[device_name] = { port: port, data_items: [] }

    REXML::XPath.each(device, ".//DataItem") do |item|
      devices[device_name][:data_items] << {
        id: item.attributes['id'],
        type: item.attributes['type'],
        category: item.attributes['category'],
        representation: item.attributes['representation'],
        values: item.get_elements("Constraints/Value").map(&:text)
      }
    end
  end
  devices
end

def generate_random_value(item)
  # 1. Пропускаем системные события ASSET
  return nil if item[:type].include?('ASSET')

  # 2. Обрабатываем CONDITION в первую очередь
  return ['normal', 'warning', 'fault'].sample if item[:category] == 'CONDITION'

  # 3. Обработка DataSet
  if item[:representation] == 'DATA_SET'
    case item[:type]
    when 'VARIABLE'
      num_pairs = rand(5..15)
      pairs = (1..num_pairs).map do
        key = rand(1..100)
        value = "%.4f" % (rand * 2000 - 1000)
        "#{key}=#{value}"
      end
      return pairs.join(' ')
    when 'SPECIFICATION_LIMIT'
      nominal = rand(50..100).to_f
      upper = nominal + rand(5..10)
      lower = nominal - rand(5..10)
      return "UPPER_LIMIT=#{upper} NOMINAL=#{nominal} LOWER_LIMIT=#{lower}"
    else
      return nil
    end
  end
  
  # 4. Проверяем предопределенные значения
  values = item[:values].empty? ? PREDEFINED_VALUES[item[:type]] : item[:values]
  return values.sample if values && !values.empty?

  # 5. Генерация по ТИПУ для всех остальных
  case item[:type]
  when 'LEVEL', 'PART_COUNT', 'TOOL_NUMBER', 'TOOL_GROUP', 'LINE_NUMBER', 'x:SEQUENCE_NUMBER'
    rand(1..100)
  when 'ROTARY_VELOCITY_OVERRIDE', 'PATH_FEEDRATE_OVERRIDE'
    rand(80..120)
  when 'POSITION', 'ANGLE', 'LENGTH'
    "%.4f" % (rand * 1000)
  when 'LOAD', 'CONCENTRATION'
    "%.2f" % (rand * 100)
  when 'ROTARY_VELOCITY', 'VELOCITY', 'ANGULAR_VELOCITY', 'CUTTING_SPEED', 'PATH_FEEDRATE', 'AXIS_FEEDRATE', 'e:SURFACE_SPEED', 'e:PATH_FEEDRATE_PER_REV'
    "%.2f" % (rand * 3000)
  when 'TEMPERATURE'
    rand(20..100)
  when 'ACCUMULATED_TIME', 'EQUIPMENT_TIMER'
    "%.2f" % (Time.now.to_f % 10000)
  when 'PATH_POSITION', 'ORIENTATION', 'TRANSLATION', 'ROTATION', 'WORK_OFFSET'
    "%.4f %.4f %.4f" % [rand * 100, rand * 100, rand * 100]
  when 'PROGRAM', 'PROGRAM_COMMENT', 'PROGRAM_HEADER', 'PROGRAM_EDIT_NAME', 'LINE_LABEL'
    ["PROG-#{rand(1000..9999)}", "Main-Op-#{rand(10..20)}", "N#{rand(100..500)}"].sample
  when 'PALLET_ID', 'FIXTURE_ID', 'TOOL_ASSET_ID', 'x:UNIT', 'PART_UNIQUE_ID', 'x:FIXTURE_UNIQUE_ID'
    "ID-#{rand(100..999)}"
  when 'MATERIAL'
    ["Aluminum-6061", "Steel-1018", "Titanium"].sample
  when 'APPLICATION', 'OPERATING_SYSTEM'
    ["SYS-#{rand(100)}", "APP-#{rand(100)}", "v#{rand(1..5)}.#{rand(0..9)}"].sample
  when 'BLOCK', 'e:BLOCK_NUMBER'
    "BLK#{rand(1..200)}"
  when 'e:MACMAN', 'e:INPUT_OUTPUT_SIGNAL'
    "SIGNAL_#{['ON', 'OFF', rand(0..1)].sample}"
  when 'x:TOOL_SUFFIX'
    "SFX-#{rand(1..9)}"
  when 'e:VARIABLES'
    "VAR_#{rand(100..999)}=%.2f" % (rand * 100)
  else
    nil
  end
end

puts "Загрузка информации об устройствах из #{DEVICES_XML_PATH}..."
devices_info = parse_devices_xml(DEVICES_XML_PATH)
puts "Найдено устройств: #{devices_info.keys.join(', ')}"

sockets = {}
servers = {}
threads = []

devices_info.each do |name, info|
    servers[name] = TCPServer.new(AGENT_HOST, info[:port])
    puts "Ожидание подключения агента к симулятору #{name} на порту #{info[:port]}..."
    threads << Thread.new(name) do |dev_name|
        loop do
            socket = nil
            begin
                socket = servers[dev_name].accept
                sockets[dev_name] = socket
                puts "Агент подключился к симулятору #{dev_name}."
                while !socket.closed? && !socket.eof?
                    socket.read_nonblock(1) rescue IO::WaitReadable
                    sleep(0.1)
                end
            rescue Errno::ECONNRESET, EOFError
            rescue => e
                puts "Неожиданная ошибка в потоке #{dev_name}: #{e.class} - #{e.message}"
            ensure
                if socket
                    puts "Соединение для #{dev_name} разорвано. Ожидание нового подключения..."
                    sockets[dev_name] = nil
                    socket.close unless socket.closed?
                end
            end
        end
    end
end

puts "\nНачинаю отправку случайных данных каждую секунду. Нажмите Ctrl+C для выхода."

begin
  loop do
    devices_info.each do |name, info|
      socket = sockets[name]
      next unless socket && !socket.closed?

      info[:data_items].each do |item|
        value = generate_random_value(item)
        next if value.nil? 
        
        timestamp = Time.now.utc.strftime("%Y-%m-%dT%H:%M:%S.%6NZ")
        line = "#{timestamp}|#{item[:id]}|#{value}"
        
        begin
          puts "[#{name}] #{line}"
          socket.puts(line)
          socket.flush
        rescue Errno::EPIPE, Errno::ECONNRESET
          break
        rescue => e
          puts "Произошла ошибка при отправке данных для #{name}: #{e.message}"
          break
        end
      end
    end
    sleep 1
  end
rescue Interrupt
  puts "\nОстановка симулятора."
ensure
  threads.each(&:kill)
  sockets.values.each { |s| s.close if s && !s.closed? }
  servers.values.each { |s| s.close if s && !s.closed? }
end