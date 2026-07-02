# Lab 0.1 — Networking Fundamentals

**Sreelakshmi K S | COGS Onboarding, InstaSafe | Week 1**

## What this lab was about

I had to set up two virtual machines that can talk to each other, capture what happens when one connects to the other over the network, look at how each machine finds its way around the network, and finally compare two ways of giving people remote access — VPN and ZTNA.

---

## Step 1: Setting up two VMs

I created two Ubuntu 22.04 VMs in VirtualBox — VM-A and VM-B. I connected them using something called an "Internal Network," basically a private little network just between these two VMs, separate from the internet connection each VM also has.

| VM | IP Address | What it does |
|---|---|---|
| VM-A | 192.168.10.1 | Client — the one making requests |
| VM-B | 192.168.10.2 | Server — the one responding |

I checked they could reach each other by pinging one from the other. Both directions worked perfectly — no packets lost at all.

---

## Step 2: Watching a connection get made

This step is about seeing exactly what happens when one computer connects to another. Every TCP connection (the kind used for things like web browsing) starts with something called a "3-way handshake" — three quick messages exchanged before any real data flows.

I wanted to use Wireshark for this, but VM-A doesn't have a desktop — it's a plain terminal-only server. So Wireshark's window couldn't open. Instead, I used a tool called `tcpdump` to capture the traffic, then read it back using `tshark` (the command-line version of Wireshark).

**How I captured it:**
```bash
sudo tcpdump -i enp0s8 -w handshake-capture.pcap port 8080 &
```
Then I made VM-A talk to VM-B's web server:
```bash
curl http://192.168.10.2:8080
```

**What I saw when I read the capture back:**

| Step | Who talks to who | What it means |
|---|---|---|
| 1 | VM-A → VM-B (SYN) | "Hey, I want to connect" |
| 2 | VM-B → VM-A (SYN, ACK) | "Okay, I'm ready, let's connect" |
| 3 | VM-A → VM-B (ACK) | "Great, we're connected now" |

After that, the actual web page request and response happened, and at the end the connection closed itself cleanly with a similar 3-step goodbye (called FIN, ACK).

---

## Step 3: Looking at the routing table

Every machine has a "routing table" — basically a list of rules for where to send traffic depending on the destination. I ran this on both VMs:
```bash
ip route show
```

**What VM-A's table told me:**

- **Default gateway:** `10.0.2.2` — this is where VM-A sends anything that isn't on its local networks (like internet traffic)
- **Its own labnet network:** `192.168.10.0/24` — this is the private network shared with VM-B. No gateway needed here since it's a direct connection.
- **Metric:** a priority number (100 on the internet route) — lower numbers get picked first when there's more than one way to reach somewhere

Here's a simple picture of how it all connects:

![Network diagram](network-diagram.png)

---

## Step 4: VPN vs ZTNA — what's the difference?

I looked into how InstaSafe's Zero Trust product (ZTAA) works and compared it to a regular VPN. Here's what stood out to me, in plain terms:

**1. How much of the network you can see**

With a VPN, once you're connected, you're basically inside the whole company network — you could technically reach lots of things unless something else blocks you. With ZTNA, you only get access to the *one specific app* you're allowed to use. Nothing else is even visible to you.

**2. How trust works**

A VPN checks you once, when you log in, and then trusts you for the whole session. ZTNA keeps checking — your identity and your device get verified continuously through something called a Controller, before you're allowed near any app.

**3. What you can see vs what stays hidden**

On a VPN, your device basically becomes part of the internal network, so you could see other systems on it. With ZTNA, the rest of the network stays completely invisible to you — you only ever see the app you're using.

![VPN vs ZTNA](ztna-vpn-diagram.png)

**Quick summary:**

| | Traditional VPN | ZTNA |
|---|---|---|
| Access | Whole network | Just one app |
| Trust check | Once, at login | Continuously |
| Visibility | You can see the network | You only see your app |
| Risk if compromised | Higher | Much lower |

*Reference: InstaSafe ZTAA documentation (docs.instasafe.com) and the internal Architecture Overview doc shared during onboarding.*
