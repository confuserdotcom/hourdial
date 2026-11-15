# Sourced by sway.sh and x11.sh (i3) after the e2e test: measures the real
# get_tree reply as the test left it and again after opening $TREE_WINDOWS
# more windows with browser-like titles, and logs the bytes each window
# adds. The caller defines wm_msg (swaymsg or i3-msg) and open_window
# (opens one window titled $1). The IPC reply cap rests on these numbers.
tree_measure() {
	go test -tags e2e ./internal/collect -run TestRealTreeSize -count=1 -v | sed -n 's/.*tree-size: //p'
}
tree_field() { printf '%s\n' "$1" | tr ' ' '\n' | sed -n "s/^$2=//p"; }

tree_n=${TREE_WINDOWS:-40}
tree_base=$(tree_measure)
echo "tree before: $tree_base"
wm_msg -q 'workspace 9' || true
i=0
while [ "$i" -lt "$tree_n" ]; do
	open_window "tock-tree-$i Pull request 1234: Lower the Sway reply cap · confuserdotcom/tock — Mozilla Firefox"
	[ "$i" -eq 0 ] && sleep 1 && { wm_msg -q 'layout tabbed' || true; }
	i=$((i + 1))
done
want=$(($(tree_field "$tree_base" windows) + tree_n))
for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
	tree_full=$(tree_measure)
	[ "$(tree_field "$tree_full" windows)" -ge "$want" ] && break
	sleep 1
done
echo "tree after:  $tree_full"
added=$(($(tree_field "$tree_full" windows) - $(tree_field "$tree_base" windows)))
grown=$(($(tree_field "$tree_full" bytes) - $(tree_field "$tree_base" bytes)))
line="tree-size ${E2E_DESKTOP:-?}: $added windows added $grown bytes, $((grown / (added > 0 ? added : 1))) bytes per window; whole tree $(tree_field "$tree_full" bytes) bytes for $(tree_field "$tree_full" windows) windows, decoding it allocated $(tree_field "$tree_full" decode_alloc) bytes"
echo "$line"
[ -n "${GITHUB_STEP_SUMMARY:-}" ] && echo "$line" >> "$GITHUB_STEP_SUMMARY"
true
