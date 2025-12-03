#!/bin/sh
set -e

# 1. 実行ファイルと static ディレクトリの存在チェック
DIR="$(cd "$(dirname "$0")" && pwd)"
if [ ! -x "${DIR}/mboxviewd" ]; then
    echo "エラー: ${DIR}/mboxviewd が見つからないか実行権限がありません。" >&2
    exit 1
fi
if [ ! -x "${DIR}/mboxappend" ]; then
    echo "エラー: ${DIR}/mboxappend が見つからないか実行権限がありません。" >&2
    exit 1
fi
if [ ! -d "${DIR}/static" ]; then
    echo "エラー: ${DIR}/static ディレクトリが見つかりません。" >&2
    exit 1
fi

# 2. バイナリを /usr/local/bin に配置
install -m 755 "${DIR}/mboxviewd" /usr/local/bin/mboxviewd
install -m 755 "${DIR}/mboxappend" /usr/local/bin/mboxappend

# 3. static ディレクトリを /usr/local/share/mboxviewd にコピー
DEST_STATIC="/usr/local/share/mboxviewd/static"
mkdir -p "$(dirname "$DEST_STATIC")"
cp -a "${DIR}/static" "$(dirname "$DEST_STATIC")"

# 4. rc.d スクリプトを生成
cat <<'EOF' > /usr/local/etc/rc.d/mboxviewd
#!/bin/sh
# PROVIDE: mboxviewd
# REQUIRE: NETWORKING
# KEYWORD: shutdown

. /etc/rc.subr

name="mboxviewd"
rcvar=mboxviewd_enable
command="/usr/local/bin/mboxviewd"

load_rc_config $name

: ${mboxviewd_user:="root"}
: ${mboxviewd_pidfile:="/var/run/mboxviewd/mboxviewd.pid"}
: ${mboxviewd_static_dir:="/usr/local/share/mboxviewd/static"}

# デフォルトフラグ（rc.conf で上書き可）
mboxviewd_flags=""

if [ -n "${mboxviewd_mbox_dir}" ]; then
    mboxviewd_flags="${mboxviewd_flags} -mbox-dir ${mboxviewd_mbox_dir}"
fi
if [ -n "${mboxviewd_static_dir}" ]; then
    mboxviewd_flags="${mboxviewd_flags} -static-dir ${mboxviewd_static_dir}"
fi
if [ -n "${mboxviewd_log_file}" ]; then
    mboxviewd_flags="${mboxviewd_flags} -log-file ${mboxviewd_log_file}"
fi
if [ -n "${mboxviewd_port}" ]; then
    mboxviewd_flags="${mboxviewd_flags} -port ${mboxviewd_port}"
fi
if [ "${mboxviewd_edit}" = "YES" ]; then
    mboxviewd_flags="${mboxviewd_flags} -edit"
fi

command_args="${mboxviewd_flags}"

start_cmd="mboxviewd_start"
stop_cmd="mboxviewd_stop"

mboxviewd_start() {
    echo "Starting ${name}..."
    su -m ${mboxviewd_user} -c "${command} ${command_args} & echo \$! > ${mboxviewd_pidfile}"
}

mboxviewd_stop() {
    if [ -f ${mboxviewd_pidfile} ]; then
        echo "Stopping ${name}..."
        kill `cat ${mboxviewd_pidfile}` && rm -f ${mboxviewd_pidfile}
    else
        echo "${name} is not running."
    fi
}

run_rc_command "$1"
EOF

chmod 755 /usr/local/etc/rc.d/mboxviewd

# 5. rc.conf にデフォルト設定を追記（既に設定が無い場合のみ）
RC_CONF="/etc/rc.conf"
{
    grep -q '^mboxviewd_enable=' "$RC_CONF" || echo 'mboxviewd_enable="YES"'
    grep -q '^mboxviewd_mbox_dir=' "$RC_CONF" || echo 'mboxviewd_mbox_dir="/usr/local/share/mail"'
    grep -q '^mboxviewd_log_file=' "$RC_CONF" || echo 'mboxviewd_log_file="/var/log/mboxviewd/mboxviewd.log"'
    grep -q '^mboxviewd_port=' "$RC_CONF" || echo 'mboxviewd_port="8080"'
    grep -q '^mboxviewd_edit=' "$RC_CONF" || echo 'mboxviewd_edit="NO"'
} >> "$RC_CONF"

# 6. サービスを有効化 & 起動
service mboxviewd enable
service mboxviewd start

echo "mboxviewd のインストールが完了しました。"
