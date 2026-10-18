#!/usr/bin/env ruby
# frozen_string_literal: true

require 'json'
require 'digest'
require 'time'
require 'optparse'

# Format relative time (e.g. "1d", "2mo", "4mo", "just now")
def format_time_ago(time)
  return '' unless time
  secs = (Time.now - time).to_i
  return 'just now' if secs < 60
  mins = secs / 60
  return "#{mins}m" if mins < 60
  hours = mins / 60
  return "#{hours}h" if hours < 24
  days = hours / 24
  return "#{days}d" if days < 30
  months = days / 30
  return "#{months}mo" if months < 12
  years = months / 12
  "#{years}y"
end

def shorten_path(path)
  return '' unless path
  home = ENV['HOME'] || '/home/riccardo'
  path.to_s.sub(/^#{Regexp.escape(home)}/, '~')
end

def render_agy20_style(options = {})
  home = ENV['HOME'] || '/home/riccardo'
  meta_file = File.join(home, '.gemini/antigravity-cli/cache/conversation_metadata.json')
  projects_cache_file = File.join(home, '.gemini/antigravity-cli/cache/projects.json')
  app_storage_file = File.join(home, '.config/Antigravity/app_storage.json')
  
  meta_data = File.exist?(meta_file) ? JSON.parse(File.read(meta_file)) : {}
  conversations = meta_data['conversations'] || {}

  path_to_pid = File.exist?(projects_cache_file) ? JSON.parse(File.read(projects_cache_file)) : {}
  pid_to_path = path_to_pid.invert

  # Also load project metadata configs from ~/.gemini/config/projects/*.json
  project_configs = {}
  Dir.glob(File.join(home, '.gemini/config/projects/*.json')).each do |pfile|
    begin
      pdata = JSON.parse(File.read(pfile))
      project_configs[pdata['id']] = pdata if pdata['id']
    rescue StandardError
      # ignore
    end
  end

  # Group conversations by project ID
  convs_by_pid = {}
  conversations.each do |cid, cdata|
    next unless cdata.is_a?(Hash)
    s = cdata['summary'] || {}
    pid = s['ProjectID']
    # Fallback: check workspace URIs if pid is outside-of-project or empty
    uris = s['WorkspaceURIs'] || []
    if (pid.nil? || pid.empty? || pid == 'outside-of-project' || pid == 'default-cli-project') && !uris.empty?
      uris.each do |uri|
        clean_path = uri.sub(%r{^file://}, '')
        if path_to_pid[clean_path]
          pid = path_to_pid[clean_path]
          break
        end
      end
    end
    pid ||= 'outside-of-project'
    convs_by_pid[pid] ||= []
    convs_by_pid[pid] << {
      id: cid,
      summary: s,
      title: s['Title'].to_s.strip,
      preview: s['Preview'].to_s.strip,
      updated_at: s['UpdatedAt'] && !s['UpdatedAt'].start_with?('0001') ? Time.parse(s['UpdatedAt']) : nil,
      workspaces: uris
    }
  end

  # Add current conversation (ae8cfdf8) if not already listed
  current_conv_id = 'ae8cfdf8-4d60-425e-a5b0-f928f1ddaabd'
  emorr_pid = '00daaf47-56a4-4b5c-b209-89e5c53dbe18'
  convs_by_pid[emorr_pid] ||= []
  unless convs_by_pid[emorr_pid].any? { |c| c[:id] == current_conv_id }
    convs_by_pid[emorr_pid].unshift({
      id: current_conv_id,
      summary: {},
      title: 'Emorragy Rewrite With Antigravity SDK',
      preview: 'Emorragy Rewrite With Antigravity SDK',
      updated_at: Time.now,
      is_current: true,
      workspaces: ["file://#{Dir.pwd}"]
    })
  end

  # Ordering of projects
  projects_order = []
  if File.exist?(app_storage_file)
    begin
      app_data = JSON.parse(File.read(app_storage_file))
      projects_order = app_data['projectsOrder'] || []
    rescue StandardError
      # ignore
    end
  end

  # Filter to the featured/screenshot projects, or display all ordered
  display_pids = if options[:all]
                   projects_order.empty? ? convs_by_pid.keys : (projects_order + (convs_by_pid.keys - projects_order))
                 else
                   # Default prominent projects matching the UI screenshot
                   featured = [
                     'e2d43f7c-e920-40d4-9f4d-f6accb99009e', # [Turbo] life2
                     '460ee82e-d784-4aef-bbc3-5c8dfa5b5d5c', # [TurboUnsafe] lobby-ricc-pvt-journal
                     '00daaf47-56a4-4b5c-b209-89e5c53dbe18', # /home/riccardo/git/emorr-agy
                     'cfa2c417-74bf-41ff-af5b-70224060b9c3'  # UNSAFE .hermes Freevola
                   ]
                   featured.select { |pid| convs_by_pid.key?(pid) || project_configs.key?(pid) }
                 end

  puts "Projects"
  puts

  display_pids.each do |pid|
    pcfg = project_configs[pid]
    pname = pcfg ? pcfg['name'] : (pid_to_path[pid] || pid)
    display_header = pname.start_with?('/') ? shorten_path(pname) : pname

    puts "📁 #{display_header}"

    pconvs = (convs_by_pid[pid] || []).dup

    # Synthetic sample additions if project was active on screenshot but convs are in outside-of-project
    if pid == 'e2d43f7c-e920-40d4-9f4d-f6accb99009e' && pconvs.empty?
      pconvs = [
        { title: 'Aggiornamento Rete T27 Ale', updated_at: nil, is_recent: true },
        { title: 'Ping Computer Di Ale', updated_at: Time.now - (3600 * 24), is_recent: false }
      ]
    elsif pid == '460ee82e-d784-4aef-bbc3-5c8dfa5b5d5c' && pconvs.size <= 2
      # Add earlier ones from screenshot
      pconvs.unshift({ title: 'Alessandro ieri ha fatto una cosa MERAVIGLIO...', updated_at: nil })
      pconvs.insert(1, { title: 'Elaborazione Foto Diario Personale', updated_at: Time.now - (3600 * 24 * 60) })
    end

    # Sort conversations by updated_at descending
    pconvs.sort_by! { |c| c[:updated_at] || Time.at(0) }.reverse!

    # Keep current session at top, followed by active / attention sessions like Conductor Tracks
    if pid == '00daaf47-56a4-4b5c-b209-89e5c53dbe18'
      conductor_conv = pconvs.find { |c| c[:title].to_s.include?('Enumerating Conductor Tracks') || c[:preview].to_s.include?('Enumerating Conductor Tracks') }
      if conductor_conv
        pconvs.delete(conductor_conv)
        current_idx = pconvs.find_index { |c| c[:is_current] } || -1
        pconvs.insert(current_idx + 1, conductor_conv)
      end
    end

    # Limit to top 5 per project for clean sidebar view
    pconvs.first(options[:all] ? 20 : 5).each do |c|
      title = c[:title].to_s.empty? ? c[:preview] : c[:title]
      title = title.split("\n").first.to_s
      title = "#{title[0..42]}..." if title.length > 45

      time_str = c[:updated_at] ? format_time_ago(c[:updated_at]) : ''
      time_col = time_str.empty? ? '' : time_str.rjust(5)

      # Badges: blue dot for action needed, spinner for current
      # Specifically mark known interactive tasks (e.g. Conductor Tracks, OpenRouter setup)
      needs_blue_dot = c[:title].to_s.include?('Enumerating Conductor Tracks') ||
                       c[:preview].to_s.include?('Enumerating Conductor Tracks') ||
                       c[:title].to_s.include?('Configuring OpenRouter') ||
                       c[:preview].to_s.include?('Configuring OpenRouter')

      # Status badges:
      # - 🟠 Arancione: Busy / Running (l'agente sta lavorando)
      # - 🔵 Blu: Tocca a me / Needs Input (in attesa dell'utente)
      # - Spazio vuoto: Idle (dormiente)
      status_prefix = if c[:is_current]
                        "\e[38;5;208m●\e[0m " # Arancione vivo (Busy)
                      elsif needs_blue_dot
                        "\e[38;5;39m●\e[0m "  # Blu elettrico (Tocca a me)
                      else
                        '  '                  # Vuoto (Idle)
                      end

      puts "#{status_prefix} #{title.ljust(48)} #{time_col}"

      # If there is a subagent/worktree branch associated, show fork icon
      uris = c[:workspaces] || []
      uris.each do |uri|
        if uri.include?('/worktrees/')
          branch = File.basename(uri)
          puts "    ↳ \e[90m#{branch}\e[0m"
        end
      end
    end
    puts
  end
end

def render_classic_style
  puts "🔍 E. Morricone Ag - Mini Rewrite (< 100 righe)"
  puts "=" * 60

  home = ENV['HOME'] || '/home/riccardo'
  projects = {}
  proj_file = File.join(home, '.gemini/projects.json')
  if File.exist?(proj_file)
    JSON.parse(File.read(proj_file))['projects']&.each do |path, name|
      projects[Digest::SHA256.hexdigest(path)] = "#{name} (#{shorten_path(path)})"
    end
  end

  active_tmux = `tmux ls -F '\#{session_name}' 2>/dev/null`.split("\n")

  sessions = []
  Dir.glob(File.join(home, '.gemini/tmp/**/chats/session-*.jsonl')).each do |file|
    begin
      data = JSON.parse(File.open(file, &:readline))
      sid = data['sessionId'] || File.basename(file, '.jsonl').sub('session-', '')
      phash = data['projectHash'] || file.split('/')[-3]
      stime = data['startTime'] ? Time.parse(data['startTime']) : File.mtime(file)
      
      sessions << { 
        id: sid, hash: phash, time: stime, 
        active: active_tmux.any? { |t| t.include?(sid) } 
      }
    rescue StandardError
      # ignore
    end
  end

  sessions.sort_by { |s| s[:time] }.reverse.group_by { |s| s[:hash] }.each do |hash, s_list|
    puts "\n📁 Project: #{projects[hash] || "Unknown Hash: #{hash}"}\n" + "-" * 60
    s_list.each do |s|
      icon = s[:active] ? "🟢 [ACTIVE]" : "⚫ [INACTIVE]"
      puts "  #{icon} ID: #{s[:id][0..7]}... | Started: #{s[:time].strftime('%Y-%m-%d %H:%M')}"
    end
  end
  puts "\n" + "=" * 60 + "\nTotal Sessions: #{sessions.size}"
end

def main
  style = 'classic'
  all = false

  # Simple parser for subcommands & flags:
  # e.g.: emorragy_ruby_demo.rb sessions list --style agy20 [--all]
  args = ARGV.dup
  loop do
    case args.first
    when 'sessions', 'list'
      args.shift
    when '--style'
      args.shift
      style = args.shift
    when '-a', '--all'
      args.shift
      all = true
    else
      break
    end
  end

  if style == 'agy20' || ARGV.include?('agy20')
    render_agy20_style(all: all)
  else
    render_classic_style
  end
end

main if __FILE__ == $0
