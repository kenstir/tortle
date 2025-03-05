# torinfo - like qbittools but for Deluge and qBittorrent

I started `torinfo` as a way to list the contents of Deluge server or qBittorrent server using Golang APIs and Cobra/Viper, and that became:

```
torinfo deluge ls --columns=ratio,name
torinfo qbit ls --columns=ratio,name
```

Then I discovered that none of the qbit "reannounce scripts" worked like I thought they should, and that became:

```
torinfo qbit reannounce [hash]
```

Maybe there will be other subcommands in the future, maybe not.

## Why another reannounce script?

Short answer: because using the autobrr builtin reannounce feature [breaks Skip Duplicates](https://discord.com/channels/881212911849209957/881967548143403058/1342160196276977725).  I would have grabbed an existing script, but thier logic didn't agree with the gold standard HBD script[^1], and none of the ones I looked at [^2],[^3],[^4],[^5] logged all the detail I needed to understand what was going on.

This implementation differs from and I think improves on the autobrr builtin reannounce in the following ways:

1. It runs separate from the autobrr hook, so the hook is free to move on and write the release status so that Skip Duplicates has a better chance of working.
2. It iterates until the tracker is ok AND there are >0 seeds.
3. it iterates 2 extra times like the deluge-reannounce script.

I plan on adding tests for (2) once I figure out how to mock the go-qbittorrent API.  But in the logs I have seen where the tracker is OK but there are no seeds.  Based on my understanding of the tracker /announce API, it seems worthwhile to keep reannouncing.   Also, I have seen where 2 trackers are OK but only the 2nd one has seeds.  It seems worthwhile in that case to stop reannouncing every 7 seconds and proceed to the extra-reannounce every 30s phase.

(3) seems the least well-reasoned, but this logic came from the HBD script for "racing using Deluge", it was duplicated in my Python fork, and it has never failed me.


[^1]: [HBD script for "racing using Deluge"](https://docs.hostingby.design/application-hosting/applications/deluge#reannounce-script).  NB: Does not work with Deluge v2.1.1, so [key-str0ke created a Python script with the same logic](https://github.com/key-str0ke/deluge-reannounce) and I [forked it and added logging]https://github.com/kenstir/deluge-reannounce/).
[^2]: [go-qbittorrent](https://github.com/autobrr/go-qbittorrent/blob/main/methods.go) func `ReannounceTorrentWithRetry`
[^3]: [qbittools](https://gitlab.com/AlexKM/qbittools/-/blob/master/commands/reannounce.py?ref_type=heads)
[^4]: [qbittorrent-cli](https://github.com/ludviglundgren/qbittorrent-cli/blob/master/cmd/torrent_reannounce.go)
[^5]: [qbtools](https://github.com/buroa/qbtools/blob/master/qbtools/commands/reannounce.py)
