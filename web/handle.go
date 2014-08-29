// Copyright © 2014 Terry Mao, LiuDing All rights reserved.
// This file is part of gopush-cluster.

// gopush-cluster is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// gopush-cluster is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with gopush-cluster.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	log "code.google.com/p/log4go"
	myrpc "github.com/Terry-Mao/gopush-cluster/rpc"
	"net/http"
	"strconv"
	"time"
)

// GetServer handle for server get
func GetServer0(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	params := r.URL.Query()
	key := params.Get("key")
	callback := params.Get("callback")
	protoStr := params.Get("proto")
	res := map[string]interface{}{"ret": OK, "msg": "ok"}
	defer retWrite(w, r, res, callback, time.Now())
	if key == "" {
		res["ret"] = ParamErr
		return
	}
	proto, err := strconv.Atoi(protoStr)
	if err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", protoStr, err)
		res["ret"] = ParamErr
		return
	}
	// Match a push-server with the value computed through ketama algorithm
	node := myrpc.GetComet(key)
	if node == nil {
		res["ret"] = NotFoundServer
		return
	}
	addrs := node.Addr[proto]
	if addrs == nil || len(addrs) == 0 {
		res["ret"] = NotFoundServer
		return
	}
	server := ""
	// Select the best ip
	if Conf.Router != "" {
		var ips []string
		for _, addr := range addrs {
			ips = append(ips, addr.Addr)
		}
		server = routerCN.SelectBest(r.RemoteAddr, ips)
		log.Debug("select the best ip:\"%s\" match with remoteAddr:\"%s\" , from ip list:\"%v\"", server, r.RemoteAddr, ips)
	}
	if server == "" {
		log.Debug("remote addr: \"%s\" chose the ip: \"%s\"", r.RemoteAddr, addrs[0].Addr)
		server = addrs[0].Addr
	}
	res["data"] = map[string]interface{}{"server": server}
	return
}

// GetOfflineMsg get offline mesage http handler.
func GetOfflineMsg0(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	params := r.URL.Query()
	key := params.Get("key")
	midStr := params.Get("mid")
	callback := params.Get("callback")
	res := map[string]interface{}{"ret": OK, "msg": "ok"}
	defer retWrite(w, r, res, callback, time.Now())
	if key == "" || midStr == "" {
		res["ret"] = ParamErr
		return
	}
	mid, err := strconv.ParseInt(midStr, 10, 64)
	if err != nil {
		res["ret"] = ParamErr
		log.Error("strconv.ParseInt(\"%s\", 10, 64) error(%v)", midStr, err)
		return
	}
	// RPC get offline messages
	reply := &myrpc.MessageGetResp{}
	args := &myrpc.MessageGetPrivateArgs{MsgId: mid, Key: key}
	client := myrpc.MessageRPC.Get()
	if client == nil {
		res["ret"] = InternalErr
		return
	}
	if err := client.Call(myrpc.MessageServiceGetPrivate, args, reply); err != nil {
		log.Error("myrpc.MessageRPC.Call(\"%s\", \"%v\", reply) error(%v)", myrpc.MessageServiceGetPrivate, args, err)
		res["ret"] = InternalErr
		return
	}
	omsgs := []string{}
	opmsgs := []string{}
	for _, msg := range reply.Msgs {
		omsg, err := msg.OldBytes()
		if err != nil {
			res["ret"] = InternalErr
			return
		}
		omsgs = append(omsgs, string(omsg))
	}

	if len(omsgs) == 0 {
		return
	}

	res["data"] = map[string]interface{}{"msgs": omsgs, "pmsgs": opmsgs}
	return
}

// GetTime get server time http handler.
func GetTime0(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	params := r.URL.Query()
	callback := params.Get("callback")
	res := map[string]interface{}{"ret": OK, "msg": "ok"}
	now := time.Now()
	defer retWrite(w, r, res, callback, now)
	res["data"] = map[string]interface{}{"timeid": now.UnixNano() / 100}
	return
}
