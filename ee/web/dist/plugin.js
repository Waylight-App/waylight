// node_modules/svelte/src/internal/disclose-version.js
typeof window < "u" && ((window.__svelte ??= {}).v ??= /* @__PURE__ */ new Set()).add("5");
// node_modules/svelte/src/constants.js
var e = {}, t = Symbol("uninitialized"), n = "http://www.w3.org/1999/xhtml", r = Array.isArray, i = Array.prototype.indexOf, a = Array.prototype.includes, o = Array.from, s = Object.defineProperty, c = Object.getOwnPropertyDescriptor, l = Object.getOwnPropertyDescriptors, u = Object.prototype, d = Array.prototype, f = Object.getPrototypeOf, p = Object.isExtensible, m = () => {};
function h(e) {
	for (var t = 0; t < e.length; t++) e[t]();
}
function g() {
	var e, t;
	return {
		promise: new Promise((n, r) => {
			e = n, t = r;
		}),
		resolve: e,
		reject: t
	};
}
var _ = 1024, v = 2048, y = 4096, b = 8192, x = 16384, S = 32768, ee = 1 << 25, C = 65536, te = 1 << 19, ne = 1 << 20, re = 1 << 25, ie = 1 << 21, ae = 1 << 22, oe = 1 << 23, se = Symbol("$state"), ce = Symbol("component"), le = Symbol(""), ue = Symbol("attributes"), de = Symbol("class"), fe = Symbol("style"), pe = Symbol("text"), me = Symbol("form reset"), he = new class extends Error {
	name = "StaleReactionError";
	message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
}(), ge = !!globalThis.document?.contentType && /* @__PURE__ */ globalThis.document.contentType.includes("xml");
function _e() {
	console.warn("https://svelte.dev/e/derived_inert");
}
function ve(e) {
	console.warn("https://svelte.dev/e/hydration_mismatch");
}
function ye() {
	console.warn("https://svelte.dev/e/svelte_boundary_reset_noop");
}
// node_modules/svelte/src/internal/client/dom/hydration.js
var w = !1;
function be(e) {
	w = e;
}
var T;
function E(t) {
	if (t === null) throw ve(), e;
	return T = t;
}
function xe() {
	return E(/* @__PURE__ */ Xt(T));
}
function Se(t) {
	if (w) {
		if (/* @__PURE__ */ Xt(T) !== null) throw ve(), e;
		T = t;
	}
}
function Ce(e = 1) {
	if (w) {
		for (var t = e, n = T; t--;) n = /* @__PURE__ */ Xt(n);
		T = n;
	}
}
function we(e = !0) {
	for (var t = 0, n = T;;) {
		if (n.nodeType === 8) {
			var r = n.data;
			if (r === "]") {
				if (t === 0) return n;
				--t;
			} else (r === "[" || r === "[!" || r[0] === "[" && !isNaN(Number(r.slice(1)))) && (t += 1);
		}
		var i = /* @__PURE__ */ Xt(n);
		e && n.remove(), n = i;
	}
}
function Te(t) {
	if (!t || t.nodeType !== 8) throw ve(), e;
	return t.data;
}
// node_modules/svelte/src/internal/client/reactivity/equality.js
function Ee(e) {
	return e === this.v;
}
function De(e, t) {
	return e == e ? e !== t || typeof e == "object" && !!e || typeof e == "function" : t == t;
}
function Oe(e) {
	return !De(e, this.v);
}
// node_modules/svelte/src/internal/client/errors.js
function ke() {
	throw Error("https://svelte.dev/e/async_derived_orphan");
}
function Ae(e, t, n) {
	throw Error("https://svelte.dev/e/each_key_duplicate");
}
function je(e) {
	throw Error("https://svelte.dev/e/effect_in_teardown");
}
function Me() {
	throw Error("https://svelte.dev/e/effect_in_unowned_derived");
}
function Ne(e) {
	throw Error("https://svelte.dev/e/effect_orphan");
}
function Pe() {
	throw Error("https://svelte.dev/e/effect_update_depth_exceeded");
}
function Fe() {
	throw Error("https://svelte.dev/e/state_descriptors_fixed");
}
function Ie() {
	throw Error("https://svelte.dev/e/state_prototype_fixed");
}
function Le() {
	throw Error("https://svelte.dev/e/state_unsafe_mutation");
}
function Re() {
	throw Error("https://svelte.dev/e/svelte_boundary_reset_onerror");
}
// node_modules/svelte/src/internal/client/context.js
var D = null;
function ze(e) {
	D = e;
}
function Be(e, t = !1, n) {
	D = {
		p: D,
		i: !1,
		c: null,
		e: null,
		s: e,
		x: null,
		r: G,
		l: null
	};
}
function Ve(e) {
	var t = D, n = t.e;
	if (n !== null) {
		t.e = null;
		for (var r of n) un(r);
	}
	return e !== void 0 && (t.x = e), t.i = !0, D = t.p, He(e);
}
function He(e = {}) {
	return s(e, ce, { value: !0 }), e;
}
function Ue() {
	return !0;
}
// node_modules/svelte/src/internal/client/dom/task.js
var We = [];
function Ge() {
	var e = We;
	We = [], h(e);
}
function Ke(e) {
	if (We.length === 0 && !yt) {
		var t = We;
		queueMicrotask(() => {
			t === We && Ge();
		});
	}
	We.push(e);
}
function qe() {
	for (; We.length > 0;) Ge();
}
// node_modules/svelte/src/internal/client/reactivity/status.js
var Je = ~(v | y | _);
function O(e, t) {
	e.f = e.f & Je | t;
}
function Ye(e) {
	e.f & 512 || e.deps === null ? O(e, _) : O(e, y);
}
// node_modules/svelte/src/internal/client/reactivity/utils.js
function Xe(e, t) {
	e.f & 2048 ? t.set(e, v) : e.f & 4096 && t.get(e) !== 2048 && t.set(e, y), O(e, _);
}
// node_modules/svelte/src/internal/client/dom/elements/misc.js
function Ze(e) {
	w && /* @__PURE__ */ Yt(e) !== null && $t(e);
}
var Qe = !1;
function $e() {
	Qe || (Qe = !0, document.addEventListener("reset", (e) => {
		Promise.resolve().then(() => {
			if (!e.defaultPrevented) for (let t of e.target.elements) t[me]?.();
		});
	}, { capture: !0 }));
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/shared.js
function et(e) {
	var t = H, n = G;
	W(null), K(null);
	try {
		return e();
	} finally {
		W(t), K(n);
	}
}
function tt(e, t, n, r = n) {
	e.addEventListener(t, () => et(n));
	let i = e[me];
	e[me] = i ? () => {
		i(), r(!0);
	} : () => r(!0), $e();
}
// node_modules/svelte/src/internal/client/reactivity/async.js
function nt(e, t, n, r) {
	let i = Ue() ? ot : ut;
	var a = e.filter((e) => !e.settled), o = t.map(i);
	if (n.length === 0 && a.length === 0) {
		r(o);
		return;
	}
	var s = G, c = rt(), l = a.length === 1 ? a[0].promise : a.length > 1 ? Promise.all(a.map((e) => e.promise)) : null;
	function u(e) {
		if (!(s.f & 16384)) {
			c();
			try {
				r([...o, ...e]);
			} catch (e) {
				L(e, s);
			}
			it();
		}
	}
	var d = at();
	if (n.length === 0) {
		l.then(() => u([])).finally(d);
		return;
	}
	function f() {
		Promise.all(n.map((e) => /* @__PURE__ */ ct(e))).then(u).catch((e) => L(e, s)).finally(d);
	}
	l ? l.then(() => {
		s.f & 16384 ? d() : (c(), f(), it());
	}) : f();
}
function rt() {
	var e = G, t = H, n = D, r = k;
	return function(i = !0) {
		K(e), W(t), ze(n), i && !(e.f & 16384) && (r?.activate(), r?.apply());
	};
}
function it(e = !0) {
	K(null), W(null), ze(null), e && k?.deactivate();
}
function at() {
	var e = G, t = e.b, n = k, r = !!t?.is_rendered();
	return t?.update_pending_count(1, n), n.increment(r, e), () => {
		t?.update_pending_count(-1, n), n.decrement(r, e);
	};
}
/*#__NO_SIDE_EFFECTS__*/
function ot(e) {
	var n = 2 | v;
	return G !== null && (G.f |= te), {
		ctx: D,
		deps: null,
		effects: null,
		equals: Ee,
		f: n,
		fn: e,
		reactions: null,
		rv: 0,
		v: t,
		wv: 0,
		parent: G,
		ac: null
	};
}
var st = Symbol("obsolete");
/*#__NO_SIDE_EFFECTS__*/
function ct(e, n, r) {
	let i = G;
	i === null && ke();
	var a = void 0, o = Ft(t), s = !H, c = /* @__PURE__ */ new Set();
	return pn(() => {
		var t = G, n = g();
		a = n.promise;
		try {
			Promise.resolve(e()).then(n.resolve, (e) => {
				e !== he && n.reject(e);
			}).finally(it);
		} catch (e) {
			n.reject(e), it();
		}
		var r = k;
		if (s) {
			if (t.f & 32768) var l = at();
			if (i.b?.is_rendered()) r.async_deriveds.get(t)?.reject(st);
			else for (let e of c.values()) e.reject(st);
			c.add(n), r.async_deriveds.set(t, n);
		}
		let u = (e, t = void 0) => {
			l?.(), c.delete(n), t !== st && (r.activate(), t ? (o.f |= oe, zt(o, t)) : (o.f & 8388608 && (o.f ^= oe), zt(o, e)), r.deactivate());
		};
		n.promise.then(u, (e) => u(null, e || "unknown"));
	}), cn(() => {
		for (let e of c) e.reject(st);
	}), new Promise((e) => {
		function t(n) {
			function r() {
				n === a ? e(o) : t(a);
			}
			n.then(r, r);
		}
		t(a);
	});
}
/*#__NO_SIDE_EFFECTS__*/
function lt(e) {
	let t = /* @__PURE__ */ ot(e);
	return kn(t), t;
}
/*#__NO_SIDE_EFFECTS__*/
function ut(e) {
	let t = /* @__PURE__ */ ot(e);
	return t.equals = Oe, t;
}
function dt(e) {
	var t = e.effects;
	if (t !== null) {
		e.effects = null;
		for (var n = 0; n < t.length; n += 1) V(t[n]);
	}
}
function ft(e) {
	var n, r = G, i = e.parent;
	if (!Dn && i !== null && e.v !== t && i.f & 24576) return _e(), e.v;
	K(i);
	try {
		dt(e), n = Rn(e);
	} finally {
		K(r);
	}
	return n;
}
function pt(e) {
	var t = ft(e);
	!e.equals(t) && (e.wv = Fn(), (!k?.is_fork || e.deps === null) && (k === null ? e.v = t : (k.capture(e, t, !0), _t?.capture(e, t, !0)), e.deps === null)) ? O(e, _) : Dn || (A === null ? Ye(e) : (sn() || k?.is_fork) && A.set(e, t));
}
function mt(e) {
	if (e.effects !== null) for (let t of e.effects) (t.teardown || t.ac) && (t.teardown?.(), t.ac !== null && et(() => {
		t.ac.abort(he), t.ac = null;
	}), t.fn !== null && (t.teardown = m), Vn(t, 0), _n(t));
}
function ht(e) {
	if (e.effects !== null) for (let t of e.effects) t.teardown && t.fn !== null && Hn(t);
}
// node_modules/svelte/src/internal/client/reactivity/batch.js
var gt = null, k = null, _t = null, A = null, vt = null, yt = !1, bt = !1, xt = null, St = null, Ct = 0, wt = 1, Tt = class e {
	id = wt++;
	started = !1;
	linked = !0;
	prev = null;
	next = null;
	async_deriveds = /* @__PURE__ */ new Map();
	current = /* @__PURE__ */ new Map();
	previous = /* @__PURE__ */ new Map();
	#e = /* @__PURE__ */ new Set();
	#t = /* @__PURE__ */ new Set();
	#n = 0;
	#r = /* @__PURE__ */ new Map();
	#i = null;
	#a = [];
	#o = [];
	#s = /* @__PURE__ */ new Map();
	#c = /* @__PURE__ */ new Map();
	#l = /* @__PURE__ */ new Set();
	is_fork = !1;
	#u = !1;
	constructor() {
		gt === null ? gt = this : (gt.next = this, this.prev = gt), gt = this;
	}
	#d() {
		if (this.is_fork) return !0;
		for (let e of this.#r.keys()) if (!this.#f(e)) return !0;
		return !1;
	}
	#f(e) {
		for (var t = e; t !== null; t = t.parent) if (this.#c.has(t)) return !0;
		return !1;
	}
	skip_effect(e) {
		this.#c.has(e) || this.#c.set(e, {
			d: [],
			m: []
		}), this.#l.delete(e);
	}
	unskip_effect(e, t = (e) => this.schedule(e)) {
		var n = this.#c.get(e);
		if (n) {
			this.#c.delete(e);
			for (var r of n.d) O(r, v), t(r);
			for (r of n.m) O(r, y), t(r);
		}
		this.#l.add(e);
	}
	#p() {
		var e = [];
		for (let i of this.#a) if (!(i.f & 16384 || !(i.f & 6144))) {
			for (var t = i, n = !1; t.parent !== null;) {
				t = t.parent;
				var r = t.f;
				if (r & 96) {
					if (!(r & 1024)) {
						n = !0;
						break;
					}
					t.f ^= _;
				}
			}
			n || e.push(t);
		}
		return this.#a = [], e;
	}
	#m() {
		this.started = !0;
		for (let [e, t] of this.#s) O(e, t), this.schedule(e);
		this.apply();
		for (var t = xt = [], n = [], r = St = []; this.#a.length > 0;) {
			Ct++ > 1e3 && (this.#b(), Dt());
			for (let e of this.#p()) try {
				this.#h(e, t, n);
			} catch (t) {
				throw Mt(e), this.#d() || this.discard(), t;
			}
		}
		if (k = null, r.length > 0) {
			var i = e.ensure();
			for (let e of r) i.schedule(e);
		}
		if (xt = null, St = null, this.#d()) {
			this.#y(n), this.#y(t);
			for (let [e, t] of this.#c) jt(e, t);
			r.length > 0 && k.#m();
			return;
		}
		let a = this.#g();
		if (a) this.#y(n), this.#y(t), a.#v(this);
		else {
			this.#s.clear();
			for (let e of this.#e) e(this);
			this.#e.clear(), _t = this, kt(n), kt(t), _t = null, this.#i?.resolve();
			var o = k;
			this.#n === 0 && (this.#a.length === 0 || o !== null) && this.#b(), o !== null && (j.clear(), o.#m());
		}
	}
	#h(e, t, n) {
		e.f ^= _;
		for (var r = e.first; r !== null;) {
			var i = r.f, a = !!(i & 96);
			if (!(a && i & 1024 || i & 8192 || this.#c.has(r)) && r.fn !== null) {
				a ? r.f ^= _ : i & 4 ? t.push(r) : In(r) && (i & 16 && this.add_dirty_reaction(r, y), Hn(r));
				var o = r.first;
				if (o !== null) {
					r = o;
					continue;
				}
			}
			for (; r !== null;) {
				var s = r.next;
				if (s !== null) {
					r = s;
					break;
				}
				r = r.parent;
			}
		}
	}
	#g() {
		for (var e = this.prev; e !== null; e = e.prev) if (!e.is_fork && this.#_(e)) return e;
		return null;
	}
	#_(e) {
		for (let [t, [, n]] of this.current) if (!n && e.current.has(t)) return !0;
		return !1;
	}
	#v(e) {
		for (let [t, n] of e.current) !this.previous.has(t) && e.previous.has(t) && this.previous.set(t, e.previous.get(t)), this.current.set(t, n);
		for (let [t, n] of e.async_deriveds) {
			let e = this.async_deriveds.get(t);
			e && n.promise.then(e.resolve).catch(e.reject);
		}
		e.async_deriveds.clear(), this.transfer_reactions(e.#s);
		let t = (e) => {
			var n = e.reactions;
			if (n !== null && !(e.f & 2 && !(e.f & 6144))) for (let e of n) {
				var r = e.f;
				if (r & 2) t(e);
				else {
					var i = e;
					r & 4194320 && !this.async_deriveds.has(i) && (this.#s.delete(i), O(i, v), this.schedule(i));
				}
			}
		};
		for (let e of this.current.keys()) t(e);
		this.oncommit(() => e.discard()), e.#b(), k = this, this.#m();
	}
	#y(e) {
		for (var t = 0; t < e.length; t += 1) Xe(e[t], this.#s);
	}
	capture(e, n, r = !1) {
		e.v !== t && !this.previous.has(e) && this.previous.set(e, e.v), e.f & 8388608 || (this.current.set(e, [n, r]), A?.set(e, n)), this.is_fork || (e.v = n);
	}
	activate() {
		k = this;
	}
	deactivate() {
		k = null, A = null;
	}
	flush() {
		try {
			bt = !0, k = this, this.#m();
		} finally {
			Ct = 0, vt = null, xt = null, St = null, bt = !1, k = null, A = null, j.clear();
		}
	}
	discard() {
		for (let e of this.#t) e(this);
		this.#t.clear();
		for (let e of this.async_deriveds.values()) e.reject(st);
		this.#b(), this.#i?.resolve();
	}
	register_created_effect(e) {
		this.#o.push(e);
	}
	increment(e, t) {
		if (this.#n += 1, e) {
			let e = this.#r.get(t) ?? 0;
			this.#r.set(t, e + 1);
		}
	}
	decrement(e, t) {
		if (--this.#n, e) {
			let e = this.#r.get(t) ?? 0;
			e === 1 ? this.#r.delete(t) : this.#r.set(t, e - 1);
		}
		this.#u || (this.#u = !0, Ke(() => {
			this.#u = !1, this.linked && this.flush();
		}));
	}
	transfer_reactions(e) {
		for (let [t, n] of e) this.add_dirty_reaction(t, n);
		e.clear();
	}
	add_dirty_reaction(e, t) {
		(t === 2048 || this.#s.get(e) !== 2048) && this.#s.set(e, t);
	}
	oncommit(e) {
		this.#e.add(e);
	}
	ondiscard(e) {
		this.#t.add(e);
	}
	settled() {
		return (this.#i ??= g()).promise;
	}
	static ensure() {
		if (k === null) {
			let t = k = new e();
			!bt && !yt && Ke(() => {
				t.started || t.flush();
			});
		}
		return k;
	}
	apply() {
		A = null;
	}
	schedule(e) {
		vt = e, e.b?.is_pending && e.f & 16777228 && !(e.f & 32768) ? e.b.defer_effect(e) : this.#a.push(e);
	}
	#b() {
		if (this.linked) {
			var e = this.prev, t = this.next;
			e === null || (e.next = t), t === null ? gt = e : t.prev = e, this.linked = !1;
		}
	}
};
function Et(e) {
	var t = yt, n = _t;
	_t = null, yt = !0;
	try {
		var r;
		for (e && (Et(), r = e());;) {
			if (qe(), k === null) return r;
			k.flush();
		}
	} finally {
		yt = t, _t = n;
	}
}
function Dt() {
	try {
		Pe();
	} catch (e) {
		L(e, vt);
	}
}
var Ot = null;
function kt(e) {
	var t = e.length;
	if (t !== 0) {
		for (var n = 0; n < t;) {
			var r = e[n++];
			if (!(r.f & 24576) && In(r) && (Ot = /* @__PURE__ */ new Set(), Hn(r), r.deps === null && r.first === null && r.nodes === null && r.teardown === null && r.ac === null && bn(r), Ot?.size > 0)) {
				j.clear();
				for (let e of Ot) {
					if (e.f & 24576) continue;
					let t = [e], n = e.parent;
					for (; n !== null;) Ot.has(n) && (Ot.delete(n), t.push(n)), n = n.parent;
					for (let e = t.length - 1; e >= 0; e--) {
						let n = t[e];
						n.f & 24576 || Hn(n);
					}
				}
				Ot.clear();
			}
		}
		Ot = null;
	}
}
function At(e) {
	k.schedule(e);
}
function jt(e, t) {
	if (!(e.f & 32 && e.f & 1024)) {
		e.f & 2048 ? t.d.push(e) : e.f & 4096 && t.m.push(e), O(e, _);
		for (var n = e.first; n !== null;) jt(n, t), n = n.next;
	}
}
function Mt(e) {
	O(e, _);
	for (var t = e.first; t !== null;) Mt(t), t = t.next;
}
// node_modules/svelte/src/internal/client/reactivity/sources.js
var Nt = /* @__PURE__ */ new Set(), j = /* @__PURE__ */ new Map(), Pt = !1;
function Ft(e, t) {
	return {
		f: 0,
		v: e,
		reactions: null,
		equals: Ee,
		rv: 0,
		wv: 0
	};
}
/*#__NO_SIDE_EFFECTS__*/
function M(e, t) {
	let n = Ft(e, t);
	return kn(n), n;
}
/*#__NO_SIDE_EFFECTS__*/
function It(e, t = !1, n = !0) {
	let r = Ft(e);
	return t || (r.equals = Oe), r;
}
function N(e, t, n = !1) {
	return H !== null && (!U || H.f & 131072) && Ue() && H.f & 4325394 && (q === null || !q.has(e)) && Le(), zt(e, n ? Ut(t) : t, St);
}
var Lt = null, Rt = 0;
function zt(e, t, n = null) {
	if (!e.equals(t)) {
		Dn ? j.set(e, t) : j.has(e) || j.set(e, e.v);
		var r = Tt.ensure();
		if (r.capture(e, t), e.f & 2) {
			let t = e;
			e.f & 2048 && ft(t), A === null && Ye(t);
		}
		e.wv = Fn(), Lt = null, Rt = 0, Ht(e, v, n), Lt = null, Ue() && G !== null && G.f & 1024 && !(G.f & 96) && (X === null ? An([e]) : X.push(e)), !r.is_fork && Nt.size > 0 && !Pt && Bt();
	}
	return t;
}
function Bt() {
	Pt = !1;
	for (let e of Nt) {
		e.f & 1024 && O(e, y);
		let t;
		try {
			t = In(e);
		} catch {
			t = !0;
		}
		t && Hn(e);
	}
	Nt.clear();
}
function Vt(e) {
	N(e, e.v + 1);
}
function Ht(e, t, n) {
	var r = e.reactions;
	if (r !== null) {
		var i = Ue(), a = r.length;
		if (Rt += a, Rt > 1e5 && Lt === null && (Lt = /* @__PURE__ */ new Set()), Lt !== null) {
			if (Lt.has(e)) return;
			Lt.add(e);
		}
		for (var o = 0; o < a; o++) {
			var s = r[o], c = s.f;
			if (i || s !== G) {
				var l = (c & v) === 0;
				if (l && O(s, t), c & 131072) Nt.add(s);
				else if (c & 2) {
					var u = s;
					A?.delete(u), Ht(u, y, n);
				} else if (l) {
					var d = s;
					c & 16 && Ot !== null && Ot.add(d), n === null ? At(d) : n.push(d);
				}
			}
		}
	}
}
function Ut(e) {
	if (typeof e != "object" || !e || se in e || ce in e) return e;
	let n = f(e);
	if (n !== u && n !== d) return e;
	var i = /* @__PURE__ */ new Map(), a = r(e), o = /* @__PURE__ */ M(0), s = null, l = Nn, p = (e) => {
		if (Nn === l) return e();
		var t = H, n = Nn;
		W(null), Pn(l);
		var r = e();
		return W(t), Pn(n), r;
	};
	return a && i.set("length", /* @__PURE__ */ M(e.length, s)), new Proxy(e, {
		defineProperty(e, t, n) {
			"value" in n && n.configurable !== !1 && n.enumerable !== !1 && n.writable !== !1 || Fe();
			var r = i.get(t);
			return r === void 0 ? p(() => {
				var e = /* @__PURE__ */ M(n.value, s);
				return i.set(t, e), e;
			}) : N(r, n.value, !0), !0;
		},
		deleteProperty(e, n) {
			var r = i.get(n);
			if (r === void 0) {
				if (n in e) {
					let e = p(() => /* @__PURE__ */ M(t, s));
					i.set(n, e), Vt(o);
				}
			} else N(r, t), Vt(o);
			return !0;
		},
		get(n, r, a) {
			if (r === se) return e;
			var o = i.get(r), l = r in n;
			if (o === void 0 && (!l || c(n, r)?.writable) && (o = p(() => /* @__PURE__ */ M(Ut(l ? n[r] : t), s)), i.set(r, o)), o !== void 0) {
				var u = Z(o);
				return u === t ? void 0 : u;
			}
			return Reflect.get(n, r, a);
		},
		getOwnPropertyDescriptor(e, n) {
			this.has?.(e, n);
			var r = Reflect.getOwnPropertyDescriptor(e, n), a = i.get(n);
			if (a !== void 0) {
				var o = Z(a);
				if (o === t) return;
				if (r && "value" in r) r.value = o;
				else return {
					enumerable: !0,
					configurable: !0,
					value: o,
					writable: !0
				};
			}
			return r;
		},
		has(e, n) {
			if (n === se) return !0;
			var r = i.get(n), a = r !== void 0 && r.v !== t || Reflect.has(e, n);
			return (r !== void 0 || (G !== null || H !== null) && (!a || c(e, n)?.writable)) && (r === void 0 && (r = p(() => /* @__PURE__ */ M(a ? Ut(e[n]) : t, s)), i.set(n, r)), Z(r) === t) ? !1 : a;
		},
		set(e, n, r, l) {
			var u = i.get(n), d = n in e;
			if (a && n === "length") for (var f = r; f < u.v; f += 1) {
				var m = i.get(f + "");
				m === void 0 ? f in e && (m = p(() => /* @__PURE__ */ M(t, s)), i.set(f + "", m)) : N(m, t);
			}
			if (u === void 0) (!d || c(e, n)?.writable) && (u = p(() => /* @__PURE__ */ M(void 0, s)), N(u, Ut(r)), i.set(n, u));
			else {
				d = u.v !== t;
				var h = p(() => Ut(r));
				N(u, h);
			}
			var g = Reflect.getOwnPropertyDescriptor(e, n);
			if (g?.set && g.set.call(l, r), !d) {
				if (a && typeof n == "string") {
					var _ = i.get("length"), v = Number(n);
					Number.isInteger(v) && v >= _.v && N(_, v + 1);
				}
				Vt(o);
			}
			return !0;
		},
		ownKeys(e) {
			Z(o);
			var n = Reflect.ownKeys(e).filter((e) => {
				var n = i.get(e);
				return n === void 0 || n.v !== t;
			});
			for (var [r, a] of i) a.v !== t && !(r in e) && n.push(r);
			return n;
		},
		setPrototypeOf() {
			Ie();
		}
	});
}
var Wt, Gt, Kt, qt;
function Jt() {
	if (Wt === void 0) {
		Wt = window, Gt = /Firefox/.test(navigator.userAgent);
		var e = Element.prototype, t = Node.prototype, n = Text.prototype;
		Kt = c(t, "firstChild").get, qt = c(t, "nextSibling").get, p(e) && (e[de] = void 0, e[ue] = null, e[fe] = void 0, e.__e = void 0), p(n) && (n[pe] = void 0);
	}
}
function P(e = "") {
	return document.createTextNode(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Yt(e) {
	return Kt.call(e);
}
/*@__NO_SIDE_EFFECTS__*/
function Xt(e) {
	return qt.call(e);
}
function Zt(e, t) {
	if (!w) return /* @__PURE__ */ Yt(e);
	var n = /* @__PURE__ */ Yt(T);
	if (n === null) n = T.appendChild(P());
	else if (t && n.nodeType !== 3) {
		var r = P();
		return n?.before(r), E(r), r;
	}
	return t && nn(n), E(n), n;
}
function Qt(e, t = !1) {
	if (!w) {
		var n = /* @__PURE__ */ Yt(e);
		return n instanceof Comment && n.data === "" ? /* @__PURE__ */ Xt(n) : n;
	}
	if (t) {
		if (T?.nodeType !== 3) {
			var r = P();
			return T?.before(r), E(r), r;
		}
		nn(T);
	}
	return T;
}
function F(e, t = !1) {
	if (!w) return /* @__PURE__ */ Yt(e);
	var n = Zt(e, t);
	return Se(e), n;
}
function I(e, t = 1, n = !1) {
	let r = w ? T : e;
	for (var i; t--;) i = r, r = /* @__PURE__ */ Xt(r);
	if (!w) return r;
	if (n) {
		if (r?.nodeType !== 3) {
			var a = P();
			return r === null ? i?.after(a) : r.before(a), E(a), a;
		}
		nn(r);
	}
	return E(r), r;
}
function $t(e) {
	e.textContent = "";
}
function en() {
	return !1;
}
function tn(e, t, n) {
	return t == null || t === "http://www.w3.org/1999/xhtml" ? n ? document.createElement(e, { is: n }) : document.createElement(e) : n ? document.createElementNS(t, e, { is: n }) : document.createElementNS(t, e);
}
function nn(e) {
	if (e.nodeValue.length < 65536) return;
	let t = e.nextSibling;
	for (; t !== null && t.nodeType === 3;) t.remove(), e.nodeValue += t.nodeValue, t = e.nextSibling;
}
function rn(e) {
	var t = G;
	if (t === null) return H.f |= oe, e;
	if (!(t.f & 32768) && !(t.f & 4)) throw e;
	L(e, t);
}
function L(t, n) {
	if (t === e) throw t;
	if (!(n !== null && n.f & 16384)) {
		for (; n !== null;) {
			if (n.f & 128 && !(n.f & 33570816)) {
				if (!(n.f & 32768)) throw t;
				try {
					n.b.error(t);
					return;
				} catch (e) {
					t = e;
				}
			}
			n = n.parent;
		}
		throw t;
	}
}
// node_modules/svelte/src/internal/client/reactivity/effects.js
function an(e) {
	G === null && (H === null && Ne(e), Me()), Dn && je(e);
}
function on(e, t) {
	var n = t.last;
	n === null ? t.last = t.first = e : (n.next = e, e.prev = n, t.last = e);
}
function R(e, t) {
	var n = G;
	n !== null && n.f & 8192 && (e |= b);
	var r = {
		ctx: D,
		deps: null,
		nodes: null,
		f: e | v | 512,
		first: null,
		fn: t,
		last: null,
		next: null,
		parent: n,
		b: n && n.b,
		prev: null,
		teardown: null,
		wv: 0,
		ac: null
	};
	k?.register_created_effect(r);
	var i = r;
	if (e & 4) xt === null ? Tt.ensure().schedule(r) : xt.push(r);
	else if (t !== null) {
		try {
			Hn(r);
		} catch (e) {
			throw V(r), e;
		}
		i.deps === null && i.teardown === null && i.nodes === null && i.first === i.last && !(i.f & 524288) && (i = i.first, e & 16 && e & 65536 && i !== null && (i.f |= C));
	}
	if (i !== null && (i.parent = n, n !== null && on(i, n), H !== null && H.f & 2 && !(e & 64))) {
		var a = H;
		(a.effects ??= []).push(i);
	}
	return r;
}
function sn() {
	return H !== null && !U;
}
function cn(e) {
	let t = R(8, null);
	return O(t, _), t.teardown = e, t;
}
function ln(e) {
	an("$effect");
	var t = G.f;
	if (!H && t & 32 && D !== null && !D.i) {
		var n = D;
		(n.e ??= []).push(e);
	} else return un(e);
}
function un(e) {
	return R(4 | ne, e);
}
function dn(e) {
	Tt.ensure();
	let t = R(64 | te, e);
	return (e = {}) => new Promise((n) => {
		e.outro ? xn(t, () => {
			V(t), n(void 0);
		}) : (V(t), n(void 0));
	});
}
function fn(e) {
	return R(4, e);
}
function pn(e) {
	return R(ae | te, e);
}
function mn(e, t = 0) {
	return R(8 | t, e);
}
function z(e, t = [], n = [], r = []) {
	nt(r, t, n, (t) => {
		R(8, () => {
			e(...t.map(Z));
		});
	});
}
function hn(e, t = 0) {
	return R(16 | t, e);
}
function B(e) {
	return R(32 | te, e);
}
function gn(e) {
	var t = e.teardown;
	if (t !== null) {
		let n = Dn, r = H;
		On(!0), W(null);
		try {
			t.call(null);
		} catch (t) {
			L(t, e.parent);
		} finally {
			On(n), W(r);
		}
	}
}
function _n(e, t = !1) {
	var n = e.first;
	for (e.first = e.last = null; n !== null;) {
		let e = n.ac;
		e !== null && et(() => {
			e.abort(he);
		});
		var r = n.next;
		n.f & 64 ? n.parent = null : V(n, t), n = r;
	}
}
function vn(e) {
	for (var t = e.first; t !== null;) {
		var n = t.next;
		t.f & 32 || V(t), t = n;
	}
}
function V(e, t = !0) {
	var n = !1;
	(t || e.f & 262144) && e.nodes !== null && e.nodes.end !== null && (yn(e.nodes.start, e.nodes.end), n = !0), e.f |= ee, _n(e, t && !n), Vn(e, 0);
	var r = e.nodes && e.nodes.t;
	if (r !== null) for (let e of r) e.stop();
	gn(e), e.f ^= ee, e.f |= x;
	var i = e.parent;
	i !== null && i.first !== null && bn(e), e.next = e.prev = e.teardown = e.ctx = e.deps = e.fn = e.nodes = e.ac = e.b = null;
}
function yn(e, t) {
	for (; e !== null;) {
		var n = e === t ? null : /* @__PURE__ */ Xt(e);
		e.remove(), e = n;
	}
}
function bn(e) {
	var t = e.parent, n = e.prev, r = e.next;
	n !== null && (n.next = r), r !== null && (r.prev = n), t !== null && (t.first === e && (t.first = r), t.last === e && (t.last = n));
}
function xn(e, t, n = !0) {
	var r = [];
	e.f |= 256, Sn(e, r, !0);
	var i = () => {
		n && V(e), t && t();
	}, a = r.length;
	if (a > 0) {
		var o = () => --a || i();
		for (var s of r) s.out(o);
	} else i();
}
function Sn(e, t, n) {
	if (!(e.f & 8192)) {
		e.f ^= b;
		var r = e.nodes && e.nodes.t;
		if (r !== null) for (let e of r) (e.is_global || n) && t.push(e);
		for (var i = e.first; i !== null;) {
			var a = i.next;
			if (!(i.f & 64)) {
				var o = !!(i.f & 65536) || !!(i.f & 32) && !!(e.f & 16);
				Sn(i, t, o ? n : !1);
			}
			i = a;
		}
	}
}
function Cn(e) {
	e.f &= -257, wn(e, !0);
}
function wn(e, t) {
	if (!(e.f & 256) && e.f & 8192) {
		e.f ^= b, e.f & 1024 || (O(e, v), Tt.ensure().schedule(e));
		for (var n = e.first; n !== null;) {
			var r = n.next, i = !!(n.f & 65536) || !!(n.f & 32);
			wn(n, i ? t : !1), n = r;
		}
		var a = e.nodes && e.nodes.t;
		if (a !== null) for (let e of a) (e.is_global || t) && e.in();
	}
}
function Tn(e, t) {
	if (e.nodes) for (var n = e.nodes.start, r = e.nodes.end; n !== null;) {
		var i = n === r ? null : /* @__PURE__ */ Xt(n);
		t.append(n), n = i;
	}
}
// node_modules/svelte/src/internal/client/legacy.js
var En = null, Dn = !1;
function On(e) {
	Dn = e;
}
var H = null, U = !1;
function W(e) {
	H = e;
}
var G = null;
function K(e) {
	G = e;
}
var q = null;
function kn(e) {
	H !== null && (H.f & 2097152 || H.f & 2) && (q ??= /* @__PURE__ */ new Set()).add(e);
}
var J = null, Y = 0, X = null;
function An(e) {
	X = e;
}
var jn = 1, Mn = 0, Nn = Mn;
function Pn(e) {
	Nn = e;
}
function Fn() {
	return ++jn;
}
function In(e) {
	var t = e.f;
	if (t & 2048) return !0;
	if (t & 4096) {
		for (var n = e.deps, r = n.length, i = 0; i < r; i++) {
			var a = n[i];
			if (In(a) && pt(a), a.wv > e.wv) return !0;
		}
		t & 512 && A === null && O(e, _);
	}
	return !1;
}
function Ln(e, t, n = !0) {
	var r = e.reactions;
	if (r !== null && !(q !== null && q.has(e))) for (var i = 0; i < r.length; i++) {
		var a = r[i];
		a.f & 2 ? Ln(a, t, !1) : t === a && (n ? O(a, v) : a.f & 1024 && O(a, y), At(a));
	}
}
function Rn(e) {
	var t = J, n = Y, r = X, i = H, a = q, o = D, s = U, c = Nn, l = e.f;
	J = null, Y = 0, X = null, H = l & 96 ? null : e, q = null, ze(e.ctx), U = !1, Nn = ++Mn, e.ac !== null && (et(() => {
		e.ac.abort(he);
	}), e.ac = null);
	try {
		e.f |= ie;
		var u = e.fn, d = u();
		e.f |= S;
		var f = zn(e);
		if (Ue() && X !== null && !U && f !== null && !(e.f & 6146)) for (var p = 0; p < X.length; p++) Ln(X[p], e);
		if (i !== null && i !== e) {
			if (Mn++, i.deps !== null) for (let e = 0; e < n; e += 1) i.deps[e].rv = Mn;
			if (t !== null) for (let e of t) e.rv = Mn;
			X !== null && (r === null ? r = X : r.push(...X));
		}
		return e.f & 8388608 && (e.f ^= oe), d;
	} catch (t) {
		return zn(e), rn(t);
	} finally {
		e.f ^= ie, J = t, Y = n, X = r, H = i, q = a, ze(o), U = s, Nn = c;
	}
}
function zn(e) {
	var t = e.deps, n = k?.is_fork;
	if (J !== null) {
		var r;
		if (n || Vn(e, Y), t !== null && Y > 0) for (t.length = Y + J.length, r = 0; r < J.length; r++) t[Y + r] = J[r];
		else e.deps = t = J;
		if (sn() && e.f & 512) for (r = Y; r < t.length; r++) (t[r].reactions ??= []).push(e);
	} else !n && t !== null && Y < t.length && (Vn(e, Y), t.length = Y);
	return t;
}
function Bn(e, n) {
	let r = n.reactions;
	if (r !== null) {
		var o = i.call(r, e);
		if (o !== -1) {
			var s = r.length - 1;
			s === 0 ? r = n.reactions = null : (r[o] = r[s], r.pop());
		}
	}
	if (r === null && n.f & 2 && (J === null || !a.call(J, n))) {
		var c = n;
		c.f & 512 && (c.f ^= 512), c.v !== t && Ye(c), c.ac !== null && et(() => {
			c.ac.abort(he), c.ac = null, O(c, v);
		}), mt(c), Vn(c, 0);
	}
}
function Vn(e, t) {
	var n = e.deps;
	if (n !== null) for (var r = t; r < n.length; r++) Bn(e, n[r]);
}
function Hn(e) {
	var t = e.f;
	if (!(t & 16384)) {
		O(e, _);
		var n = G;
		G = e;
		try {
			t & 16777232 ? vn(e) : _n(e), gn(e);
			var r = Rn(e);
			e.teardown = typeof r == "function" ? r : null, e.wv = jn;
		} finally {
			G = n;
		}
	}
}
async function Un() {
	await Promise.resolve(), Et();
}
function Z(e) {
	var t = !!(e.f & 2);
	if (En?.add(e), H !== null && !U && !(G !== null && G.f & 16384) && (q === null || !q.has(e))) {
		var n = H.deps;
		if (H.f & 2097152) e.rv < Mn && (e.rv = Mn, J === null && n !== null && n[Y] === e ? Y++ : J === null ? J = [e] : J.push(e));
		else {
			H.deps ??= [], a.call(H.deps, e) || H.deps.push(e);
			var r = e.reactions;
			r === null ? e.reactions = [H] : a.call(r, H) || r.push(H);
		}
	}
	if (Dn && j.has(e)) return j.get(e);
	if (t) {
		var i = e;
		if (Dn) {
			var o = i.v;
			return (!(i.f & 1024) && i.reactions !== null || Gn(i)) && (o = ft(i)), j.set(i, o), o;
		}
		var s = !(i.f & 512) && !U && H !== null && !!(H.f & 512), c = (i.f & S) === 0;
		In(i) && (s && (i.f |= 512), pt(i)), s && !c && (ht(i), Wn(i));
	}
	if (A?.has(e)) return A.get(e);
	if (e.f & 8388608) throw e.v;
	return e.v;
}
function Wn(e) {
	if (e.f |= 512, e.deps !== null) for (let n of e.deps) {
		var t = n.reactions;
		t === null ? n.reactions = [e] : a.call(t, e) || t.push(e), n.f & 2 && !(n.f & 512) && (ht(n), Wn(n));
	}
}
function Gn(e) {
	if (e.v === t) return !0;
	if (e.deps === null) return !1;
	for (let t of e.deps) if (j.has(t) || t.f & 2 && Gn(t)) return !0;
	return !1;
}
function Kn(e) {
	var t = U;
	try {
		return U = !0, e();
	} finally {
		U = t;
	}
}
// node_modules/svelte/src/internal/client/dom/elements/events.js
var qn = Symbol("events"), Jn = /* @__PURE__ */ new Set(), Yn = /* @__PURE__ */ new Set();
function Xn(e, t, n, r = {}) {
	function i(e) {
		if (r.capture || nr.call(t, e), !e.cancelBubble) return et(() => n?.call(this, e));
	}
	return e.startsWith("pointer") || e.startsWith("touch") || e === "wheel" ? (i.__removed = !1, Ke(() => {
		i.__removed || t.addEventListener(e, i, r);
	})) : t.addEventListener(e, i, r), i;
}
function Zn(e, t, n, r, i) {
	var a = {
		capture: r,
		passive: i
	}, o = Xn(e, t, n, a);
	(t === document.body || t === window || t === document || t instanceof HTMLMediaElement) && cn(() => {
		o.__removed = !0, t.removeEventListener(e, o, a);
	});
}
function Qn(e, t, n) {
	(t[qn] ??= {})[e] = n;
}
function $n(e) {
	for (var t = 0; t < e.length; t++) Jn.add(e[t]);
	for (var n of Yn) n(e);
}
var er = null, tr = !1;
function nr(e) {
	var t = this, n = t.ownerDocument, r = e.type, i = e.composedPath?.() || [], a = i[0] || e.target;
	er = e, tr || (tr = !0, setTimeout(() => {
		tr = !1, er = null;
	}));
	var o = 0, c = er === e && e[qn];
	if (c) {
		var l = i.indexOf(c);
		if (l !== -1 && (t === document || t === window)) {
			e[qn] = t;
			return;
		}
		var u = i.indexOf(t);
		if (u === -1) return;
		l <= u && (o = l);
	}
	if (a = i[o] || e.target, a !== t) {
		s(e, "currentTarget", {
			configurable: !0,
			get() {
				return a || n;
			}
		});
		var d = H, f = G;
		W(null), K(null);
		try {
			for (var p, m = []; a !== null && a !== t;) {
				try {
					var h = a[qn]?.[r];
					h == null || a.disabled && e.target !== a || h.call(a, e);
				} catch (e) {
					p ? m.push(e) : p = e;
				}
				if (e.cancelBubble) break;
				o++, a = o < i.length ? i[o] : null;
			}
			if (p) {
				for (let e of m) queueMicrotask(() => {
					throw e;
				});
				throw p;
			}
		} finally {
			e[qn] = t, delete e.currentTarget, W(d), K(f);
		}
	}
}
// node_modules/svelte/src/internal/client/dom/reconciler.js
var rr = globalThis?.window?.trustedTypes && /* @__PURE__ */ globalThis.window.trustedTypes.createPolicy("svelte-trusted-html", { createHTML: (e) => e });
function ir(e) {
	return rr?.createHTML(e) ?? e;
}
function ar(e) {
	var t = tn("template");
	return t.innerHTML = ir(e.replaceAll("<!>", "<!---->")), t.content;
}
// node_modules/svelte/src/internal/client/dom/template.js
function or(e, t) {
	var n = G;
	n.nodes === null && (n.nodes = {
		start: e,
		end: t,
		a: null,
		t: null
	});
}
/*#__NO_SIDE_EFFECTS__*/
function sr(e, t) {
	var n = !!(t & 1), r = !!(t & 2), i, a = !e.startsWith("<!>");
	return () => {
		if (w) return or(T, null), T;
		i === void 0 && (i = ar(a ? e : "<!>" + e), n || (i = /* @__PURE__ */ Yt(i)));
		var t = r || Gt ? document.importNode(i, !0) : i.cloneNode(!0);
		if (n) {
			var o = /* @__PURE__ */ Yt(t), s = t.lastChild;
			or(o, s);
		} else or(t, t);
		return t;
	};
}
function Q(e, t) {
	if (w) {
		var n = G;
		n.f & 32768 && n.nodes.end !== null || (n.nodes.end = T), xe();
	} else e !== null && e.before(t);
}
[.../* @__PURE__ */ "allowfullscreen.async.autofocus.autoplay.checked.controls.default.disabled.formnovalidate.indeterminate.inert.ismap.loop.multiple.muted.nomodule.novalidate.open.playsinline.readonly.required.reversed.seamless.selected.webkitdirectory.defer.disablepictureinpicture.disableremoteplayback".split(".")];
var cr = ["touchstart", "touchmove"];
function lr(e) {
	return cr.includes(e);
}
// node_modules/svelte/src/reactivity/create-subscriber.js
function ur(e) {
	let t = 0, n = Ft(0), r;
	return () => {
		sn() && (Z(n), mn(() => (t === 0 && (r = Kn(() => e(() => Vt(n)))), t += 1, () => {
			Ke(() => {
				--t, t === 0 && (r?.(), r = void 0, Vt(n));
			});
		})));
	};
}
// node_modules/svelte/src/internal/client/dom/blocks/boundary.js
var dr = C | te;
function fr(e, t, n, r) {
	new pr(e, t, n, r);
}
var pr = class {
	parent;
	is_pending = !1;
	transform_error;
	#e;
	#t = w ? T : null;
	#n;
	#r;
	#i;
	#a = null;
	#o = null;
	#s = null;
	#c = null;
	#l = 0;
	#u = 0;
	#d = !1;
	#f = /* @__PURE__ */ new Map();
	#p = null;
	#m = ur(() => (this.#p = Ft(this.#l), () => {
		this.#p = null;
	}));
	constructor(e, t, n, r) {
		this.#e = e, this.#n = t, this.#r = (e) => {
			var t = G;
			t.b = this, t.f |= 128, n(e);
		}, this.parent = G.b, this.transform_error = r ?? this.parent?.transform_error ?? ((e) => e), this.#i = hn(() => {
			if (w) {
				let e = Te(this.#t);
				xe();
				let t = e === "[!";
				if (e.startsWith("[?")) {
					let t = JSON.parse(e.slice(2));
					this.#g(t);
				} else t ? this.#y() : this.#h();
			} else this.#b();
		}, dr), w && (this.#e = T);
	}
	#h() {
		try {
			this.#a = B(() => this.#r(this.#e));
		} catch (e) {
			this.error(e);
		}
	}
	#g(e) {
		let t = this.#n.failed, { reset: n, invoke_onerror: r } = this.#_(e);
		Ke(r), t && (this.#s = B(() => {
			t(this.#e, () => e, () => n);
		}));
	}
	#_(e) {
		var t = !1, n = !1;
		let r = () => {
			this.#v() || (t ? ye() : (t = !0, n && Re(), this.#s !== null && xn(this.#s, () => {
				this.#s = null;
			}), this.#S(() => {
				this.#b();
			})));
		};
		return {
			reset: r,
			invoke_onerror: () => {
				if (!this.#v()) try {
					n = !0, this.#n.onerror?.(e, r), n = !1;
				} catch (e) {
					L(e, this.#i && this.#i.parent);
				}
			}
		};
	}
	#v() {
		return (this.#i.f & (x | ee)) !== 0;
	}
	#y() {
		let e = this.#n.pending;
		e && (this.is_pending = !0, this.#o = B(() => e(this.#e)), Ke(() => {
			if (!this.#v()) {
				var e = this.#c = document.createDocumentFragment(), t = P(), n = !1;
				e.append(t), this.#a = this.#S(() => {
					try {
						return B(() => this.#r(t));
					} catch (e) {
						try {
							this.error(e), n = !0;
						} catch (e) {
							L(e, this.#i.parent);
						}
						return null;
					}
				}), this.#a === null ? (this.#c = null, n && this.#x(k)) : this.#u === 0 && (this.#e.before(e), this.#c = null, xn(this.#o, () => {
					this.#o = null;
				}), this.#x(k));
			}
		}));
	}
	#b() {
		try {
			if (this.is_pending = this.has_pending_snippet(), this.#u = 0, this.#l = 0, this.#a = B(() => {
				this.#r(this.#e);
			}), this.#u > 0) {
				var e = this.#c = document.createDocumentFragment();
				Tn(this.#a, e);
				let t = this.#n.pending;
				this.#o = B(() => t(this.#e));
			} else this.#x(k);
		} catch (e) {
			this.error(e);
		}
	}
	#x(e) {
		this.is_pending = !1, e.transfer_reactions(this.#f);
	}
	defer_effect(e) {
		Xe(e, this.#f);
	}
	is_rendered() {
		return !this.is_pending && (!this.parent || this.parent.is_rendered());
	}
	has_pending_snippet() {
		return !!this.#n.pending;
	}
	#S(e) {
		var t = G, n = H, r = D;
		K(this.#i), W(this.#i), ze(this.#i.ctx);
		try {
			return Tt.ensure(), e();
		} finally {
			K(t), W(n), ze(r);
		}
	}
	#C(e, t) {
		this.has_pending_snippet() ? (this.#u += e, this.#u === 0 && (this.#x(t), this.#o && xn(this.#o, () => {
			this.#o = null;
		}), this.#c &&= (this.#e.before(this.#c), null))) : this.parent && this.parent.#C(e, t);
	}
	update_pending_count(e, t) {
		this.#C(e, t), this.#l += e, !this.#p || this.#d || (this.#d = !0, Ke(() => {
			this.#d = !1, this.#p && zt(this.#p, this.#l);
		}));
	}
	get_effect_pending() {
		return this.#m(), Z(this.#p);
	}
	error(t) {
		if (t === e || !this.#n.onerror && !this.#n.failed) throw t;
		k?.is_fork ? (this.#a && k.skip_effect(this.#a), this.#o && k.skip_effect(this.#o), this.#s && k.skip_effect(this.#s), k.oncommit(() => {
			this.#v() || this.#w(t);
		})) : this.#w(t);
	}
	#w(e) {
		this.#a &&= (V(this.#a), null), this.#o &&= (V(this.#o), null), this.#s &&= (V(this.#s), null), w && (E(this.#t), Ce(), E(we()));
		let t = this.#n.failed, n = (e) => {
			if (this.#v()) return;
			let { reset: n, invoke_onerror: r } = this.#_(e);
			r(), t && !this.#v() && (this.#s = this.#S(() => {
				try {
					return B(() => {
						var r = G;
						r.b = this, r.f |= 128, t(this.#e, () => e, () => n);
					});
				} catch (e) {
					return L(e, this.#i.parent), null;
				}
			}));
		};
		Ke(() => {
			if (!this.#v()) {
				var t;
				try {
					t = this.transform_error(e);
				} catch (e) {
					L(e, this.#i && this.#i.parent);
					return;
				}
				typeof t == "object" && t && typeof t.then == "function" ? t.then(n, (e) => L(e, this.#i && this.#i.parent)) : n(t);
			}
		});
	}
};
function $(e, t) {
	var n = t == null ? "" : typeof t == "object" ? `${t}` : t;
	n !== (e[pe] ??= e.nodeValue) && (e[pe] = n, e.nodeValue = `${n}`);
}
function mr(e, t) {
	return gr(e, t);
}
var hr = /* @__PURE__ */ new Map();
function gr(t, { target: n, anchor: r, props: i = {}, events: a, context: s, intro: c = !0, transformError: l }) {
	Jt();
	var u = void 0, d = dn(() => {
		var c = r ?? n.appendChild(P());
		fr(c, { pending: () => {} }, (n) => {
			Be({});
			var r = D;
			if (s && (r.c = s), a && (i.$$events = a), w && or(n, null), u = t(n, i) || He(), w && (G.nodes.end = T, T === null || T.nodeType !== 8 || T.data !== "]")) throw ve(), e;
			Ve();
		}, l);
		var d = /* @__PURE__ */ new Set(), f = (e) => {
			for (var t = 0; t < e.length; t++) {
				var r = e[t];
				if (!d.has(r)) {
					d.add(r);
					var i = lr(r);
					for (let e of [n, document]) {
						var a = hr.get(e);
						a === void 0 && (a = /* @__PURE__ */ new Map(), hr.set(e, a));
						var o = a.get(r);
						o === void 0 ? (e.addEventListener(r, nr, { passive: i }), a.set(r, 1)) : a.set(r, o + 1);
					}
				}
			}
		};
		return f(o(Jn)), Yn.add(f), () => {
			for (var e of d) for (let r of [n, document]) {
				var t = hr.get(r), i = t.get(e);
				--i == 0 ? (r.removeEventListener(e, nr), t.delete(e), t.size === 0 && hr.delete(r)) : t.set(e, i);
			}
			Yn.delete(f), c !== r && c.parentNode?.removeChild(c);
		};
	});
	return _r.set(u, d), u;
}
var _r = /* @__PURE__ */ new WeakMap();
function vr(e, t) {
	let n = _r.get(e);
	return n ? (_r.delete(e), n(t)) : Promise.resolve();
}
// node_modules/svelte/src/internal/client/dom/blocks/branches.js
var yr = class {
	anchor;
	#e = /* @__PURE__ */ new Map();
	#t = /* @__PURE__ */ new Map();
	#n = /* @__PURE__ */ new Map();
	#r = /* @__PURE__ */ new Set();
	#i = !0;
	constructor(e, t = !0) {
		this.anchor = e, this.#i = t;
	}
	#a = (e) => {
		if (this.#e.has(e)) {
			var t = this.#e.get(e), n = this.#t.get(t);
			if (n) Cn(n), this.#r.delete(t);
			else {
				var r = this.#n.get(t);
				r && (Cn(r.effect), this.#t.set(t, r.effect), this.#n.delete(t), r.fragment.lastChild.remove(), this.anchor.before(r.fragment), n = r.effect);
			}
			for (let [t, n] of this.#e) {
				if (this.#e.delete(t), t === e) break;
				let r = this.#n.get(n);
				r && (V(r.effect), this.#n.delete(n));
			}
			for (let [e, r] of this.#t) {
				if (e === t || this.#r.has(e)) continue;
				let i = () => {
					if (Array.from(this.#e.values()).includes(e)) {
						var t = document.createDocumentFragment();
						Tn(r, t), t.append(P()), this.#n.set(e, {
							effect: r,
							fragment: t
						});
					} else V(r);
					this.#r.delete(e), this.#t.delete(e);
				};
				this.#i || !n ? (this.#r.add(e), xn(r, i, !1)) : i();
			}
		}
	};
	#o = (e) => {
		this.#e.delete(e);
		let t = Array.from(this.#e.values());
		for (let [e, n] of this.#n) t.includes(e) || (V(n.effect), this.#n.delete(e));
	};
	ensure(e, t) {
		var n = k, r = en();
		if (t && !this.#t.has(e) && !this.#n.has(e)) {
			if (r) {
				var i = document.createDocumentFragment(), a = P();
				i.append(a), this.#n.set(e, {
					effect: B(() => t(a)),
					fragment: i
				});
			} else this.#t.set(e, B(() => t(this.anchor)));
		}
		if (this.#e.set(n, e), r) {
			for (let [t, r] of this.#t) t === e ? n.unskip_effect(r) : n.skip_effect(r);
			for (let [t, r] of this.#n) t === e ? n.unskip_effect(r.effect) : n.skip_effect(r.effect);
			n.oncommit(this.#a), n.ondiscard(this.#o);
		} else w && (this.anchor = T), this.#a(n);
	}
};
// node_modules/svelte/src/internal/client/dom/blocks/if.js
function br(e, t, n = !1) {
	var r;
	w && (r = T, xe());
	var i = new yr(e), a = n ? C : 0;
	function o(e, t) {
		if (w) {
			var n = Te(r);
			if (e !== parseInt(n.substring(1))) {
				var a = we();
				E(a), i.anchor = a, be(!1), i.ensure(e, t), be(!0);
				return;
			}
		}
		i.ensure(e, t);
	}
	hn(() => {
		var e = !1;
		t((t, n = 0) => {
			e = !0, o(n, t);
		}), e || o(-1, null);
	}, a);
}
// node_modules/svelte/src/internal/client/dom/blocks/each.js
function xr(e, t, n) {
	for (var r = [], i = t.length, a, s = t.length, c = 0; c < i; c++) {
		let n = t[c];
		xn(n, () => {
			if (a) {
				if (a.pending.delete(n), a.done.add(n), a.pending.size === 0) {
					var t = e.outrogroups;
					Sr(e, o(a.done)), t.delete(a), t.size === 0 && (e.outrogroups = null);
				}
			} else --s;
		}, !1);
	}
	if (s === 0) {
		var l = r.length === 0 && n !== null && e.pending.size === 0;
		if (l) {
			var u = n, d = u.parentNode;
			$t(d), d.append(u), e.items.clear();
		}
		Sr(e, t, !l);
	} else a = {
		pending: new Set(t),
		done: /* @__PURE__ */ new Set()
	}, (e.outrogroups ??= /* @__PURE__ */ new Set()).add(a);
}
function Sr(e, t, n = !0) {
	var r;
	if (e.pending.size > 0) {
		r = /* @__PURE__ */ new Set();
		for (let t of e.pending.values()) for (let n of t) r.add(e.items.get(n).e);
	}
	for (var i = 0; i < t.length; i++) {
		var a = t[i];
		r?.has(a) ? (a.f |= re, Tn(a, document.createDocumentFragment())) : V(t[i], n);
	}
}
var Cr;
function wr(e, t, n, i, a, s = null) {
	var c = e, l = /* @__PURE__ */ new Map();
	if (t & 4) {
		var u = e;
		c = w ? E(/* @__PURE__ */ Yt(u)) : u.appendChild(P());
	}
	w && xe();
	var d = null, f = /* @__PURE__ */ ut(() => {
		var e = n();
		return r(e) ? e : e == null ? [] : o(e);
	}), p = /* @__PURE__ */ new Map(), m = !0;
	function h(e) {
		if (!(_.effect.f & 16384)) {
			_.pending.delete(e);
			var n = Z(f);
			_.fallback = d, Er(_, n, c, t, i), d !== null && (n.length === 0 ? d.f & 33554432 ? (d.f ^= re, Or(d, null, c)) : Cn(d) : xn(d, () => {
				d = null;
			}));
		}
	}
	function g(e) {
		_.pending.delete(e);
	}
	var _ = {
		effect: hn(() => {
			var e = Z(f), r = e.length;
			let o = !1;
			w && Te(c) === "[!" != (r === 0) && (c = we(), E(c), be(!1), o = !0);
			for (var u = /* @__PURE__ */ new Set(), _ = k, v = en(), y = 0; y < r; y += 1) {
				w && T.nodeType === 8 && T.data === "]" && (c = T, o = !0, be(!1));
				var b = e[y], x = i(b, y), S = m ? null : l.get(x);
				S ? (S.v && zt(S.v, b), S.i && zt(S.i, y), v && _.unskip_effect(S.e)) : (S = Dr(l, m ? c : Cr ??= P(), b, x, y, a, t, n), m || (S.e.f |= re), l.set(x, S)), u.add(x);
			}
			if (r === 0 && s && !d && (m ? d = B(() => s(c)) : (d = B(() => s(Cr ??= P())), d.f |= re)), r > u.size && Ae("", "", ""), w && r > 0 && E(we()), !m) {
				if (p.set(_, u), v) {
					for (let [e, t] of l) u.has(e) || _.skip_effect(t.e);
					_.oncommit(h), _.ondiscard(g);
				} else h(_);
			}
			o && be(!0), Z(f);
		}),
		flags: t,
		items: l,
		pending: p,
		outrogroups: null,
		fallback: d
	};
	m = !1, w && (c = T);
}
function Tr(e) {
	for (; e !== null && !(e.f & 32);) e = e.next;
	return e;
}
function Er(e, t, n, r, i) {
	var a = !!(r & 8), s = t.length, c = e.items, l = Tr(e.effect.first), u, d = null, f, p = [], m = [], h, g, _, v;
	if (a) for (v = 0; v < s; v += 1) h = t[v], g = i(h, v), _ = c.get(g).e, _.f & 33554432 || (_.nodes?.a?.measure(), (f ??= /* @__PURE__ */ new Set()).add(_));
	for (v = 0; v < s; v += 1) {
		if (h = t[v], g = i(h, v), _ = c.get(g).e, e.outrogroups !== null) for (let t of e.outrogroups) t.pending.delete(_), t.done.delete(_);
		if (_.f & 8192 && (Cn(_), a && (_.nodes?.a?.unfix(), (f ??= /* @__PURE__ */ new Set()).delete(_))), _.f & 33554432) {
			if (_.f ^= re, _ === l) Or(_, null, n);
			else {
				var y = d ? d.next : l;
				_ === e.effect.last && (e.effect.last = _.prev), _.prev && (_.prev.next = _.next), _.next && (_.next.prev = _.prev), kr(e, d, _), kr(e, _, y), Or(_, y, n), d = _, p = [], m = [], l = Tr(d.next);
				continue;
			}
		}
		if (_ !== l) {
			if (u !== void 0 && u.has(_)) {
				if (p.length < m.length) {
					var b = m[0], x;
					d = b.prev;
					var S = p[0], ee = p[p.length - 1];
					for (x = 0; x < p.length; x += 1) Or(p[x], b, n);
					for (x = 0; x < m.length; x += 1) u.delete(m[x]);
					kr(e, S.prev, ee.next), kr(e, d, S), kr(e, ee, b), l = b, d = ee, --v, p = [], m = [];
				} else u.delete(_), Or(_, l, n), kr(e, _.prev, _.next), kr(e, _, d === null ? e.effect.first : d.next), kr(e, d, _), d = _;
				continue;
			}
			for (p = [], m = []; l !== null && l !== _;) (u ??= /* @__PURE__ */ new Set()).add(l), m.push(l), l = Tr(l.next);
			if (l === null) continue;
		}
		_.f & 33554432 || p.push(_), d = _, l = Tr(_.next);
	}
	if (e.outrogroups !== null) {
		for (let t of e.outrogroups) t.pending.size === 0 && (Sr(e, o(t.done)), e.outrogroups?.delete(t));
		e.outrogroups.size === 0 && (e.outrogroups = null);
	}
	if (l !== null || u !== void 0) {
		var C = [];
		if (u !== void 0) for (_ of u) _.f & 8192 || C.push(_);
		for (; l !== null;) !(l.f & 8192) && l !== e.fallback && C.push(l), l = Tr(l.next);
		var te = C.length;
		if (te > 0) {
			var ne = r & 4 && s === 0 ? n : null;
			if (a) {
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.measure();
				for (v = 0; v < te; v += 1) C[v].nodes?.a?.fix();
			}
			xr(e, C, ne);
		}
	}
	a && Ke(() => {
		if (f !== void 0) for (_ of f) _.nodes?.a?.apply();
	});
}
function Dr(e, t, n, r, i, a, o, s) {
	var c = o & 1 ? o & 16 ? Ft(n) : /* @__PURE__ */ It(n, !1, !1) : null, l = o & 2 ? Ft(i) : null;
	return {
		v: c,
		i: l,
		e: B(() => (a(t, c ?? n, l ?? i, s), () => {
			e.delete(r);
		}))
	};
}
function Or(e, t, n) {
	if (e.nodes) for (var r = e.nodes.start, i = e.nodes.end, a = t && !(t.f & 33554432) ? t.nodes.start : n; r !== null;) {
		var o = /* @__PURE__ */ Xt(r);
		if (a.before(r), r === i) return;
		r = o;
	}
}
function kr(e, t, n) {
	t === null ? e.effect.first = n : t.next = n, n === null ? e.effect.last = t : n.prev = t;
}
// node_modules/svelte/src/internal/client/dom/css.js
function Ar(e, t) {
	fn(() => {
		e = G?.parent?.nodes?.start ?? e;
		var n = e.getRootNode(), r = n.host ? n : n.head ?? n.ownerDocument.head;
		if (!r.querySelector("#" + t.hash)) {
			let e = tn("style");
			e.id = t.hash, e.textContent = t.code, r.appendChild(e);
		}
	});
}
// node_modules/svelte/src/internal/shared/attributes.js
var jr = [..." 	\n\r\f\xA0\v﻿"];
function Mr(e, t, n) {
	var r = e == null ? "" : "" + e;
	if (t && (r = r ? r + " " + t : t), n) {
		for (var i of Object.keys(n)) if (n[i]) r = r ? r + " " + i : i;
		else if (r.length) for (var a = i.length, o = 0; (o = r.indexOf(i, o)) >= 0;) {
			var s = o + a;
			(o === 0 || jr.includes(r[o - 1])) && (s === r.length || jr.includes(r[s])) ? r = (o === 0 ? "" : r.substring(0, o)) + r.substring(s + 1) : o = s;
		}
	}
	return r === "" ? null : r;
}
// node_modules/svelte/src/internal/client/dom/elements/class.js
function Nr(e, t, n, r, i, a) {
	var o = e[de];
	if (w || o !== n || o === void 0) {
		var s = Mr(n, r, a);
		w && s === e.getAttribute("class") || (s == null ? e.removeAttribute("class") : t ? e.className = s : e.setAttribute("class", s)), e[de] = n;
	} else if (a && i !== a) for (var c in a) {
		var l = !!a[c];
		(i == null || l !== !!i[c]) && e.classList.toggle(c, l);
	}
	return a;
}
// node_modules/svelte/src/internal/client/dom/elements/attributes.js
var Pr = Symbol("is custom element"), Fr = Symbol("is html"), Ir = ge ? "link" : "LINK";
function Lr(e, t, n, r) {
	var i = Rr(e);
	if (!(w && (i[t] = e.getAttribute(t), t === "src" || t === "srcset" || t === "href" && e.nodeName === Ir))) {
		var a = i[t];
		a !== (i[t] = n) && (t === "loading" && (e[le] = n), n == null ? e.removeAttribute(t) : typeof n != "string" && Br(e).has(t) ? e[t] = n : (!w || typeof n != "number" && typeof n != "boolean" || a !== String(n)) && e.setAttribute(t, n));
	}
}
function Rr(e) {
	return e[ue] ??= {
		[Pr]: e.nodeName.includes("-"),
		[Fr]: e.namespaceURI === n
	};
}
var zr = /* @__PURE__ */ new Map();
function Br(e) {
	var t = e.getAttribute("is") || e.nodeName, n = zr.get(t);
	if (n) return n;
	zr.set(t, n = /* @__PURE__ */ new Set());
	for (var r, i = e, a = Element.prototype; a !== i;) {
		for (var o in r = l(i), r) r[o].set && o !== "innerHTML" && o !== "textContent" && o !== "innerText" && n.add(o);
		i = f(i);
	}
	return n;
}
// node_modules/svelte/src/internal/client/dom/elements/bindings/input.js
function Vr(e, t, n = t) {
	var r = /* @__PURE__ */ new WeakSet();
	tt(e, "input", async (i) => {
		var a = i ? e.defaultValue : e.value;
		if (a = Hr(e) ? Ur(a) : a, n(a), k !== null && r.add(k), await Un(), a !== (a = t())) {
			var o = e.selectionStart, s = e.selectionEnd, c = e.value.length;
			if (e.value = a ?? "", s !== null) {
				var l = e.value.length;
				o === s && s === c && l > c ? (e.selectionStart = l, e.selectionEnd = l) : (e.selectionStart = o, e.selectionEnd = Math.min(s, l));
			}
		}
	}), (w && e.defaultValue !== e.value || Kn(t) == null && e.value) && (n(Hr(e) ? Ur(e.value) : e.value), k !== null && r.add(k)), mn(() => {
		var n = t();
		if (e === document.activeElement) {
			var i = k;
			if (r.has(i)) return;
		}
		Hr(e) && n === Ur(e.value) || (e.type !== "date" || n || e.value) && n !== e.value && (e.value = n ?? "");
	});
}
function Hr(e) {
	var t = e.type;
	return t === "number" || t === "range";
}
function Ur(e) {
	return e === "" ? null : +e;
}
// src/lib/texts.ts
function Wr(e, t, n, r = {}) {
	let i = e[t];
	return (i ? i[n] ?? i.en : t).replace(/\{(\w+)\}/g, (e, t) => t in r ? String(r[t]) : e);
}
// src/licence/features.ts
var Gr = [
	{
		id: "sso",
		name: {
			en: "Single sign-on for all accounts",
			de: "Single Sign-on für alle Konten"
		},
		description: {
			en: "Members sign in with your identity provider; groups are mapped.",
			de: "Mitglieder melden sich über euren Identity Provider an; Gruppen werden übernommen."
		}
	},
	{
		id: "directory-sync",
		name: {
			en: "Directory sync",
			de: "Verzeichnis-Abgleich"
		},
		description: {
			en: "Accounts and groups from LDAP, Active Directory or SCIM.",
			de: "Konten und Gruppen aus LDAP, Active Directory oder SCIM."
		}
	},
	{
		id: "policies",
		name: {
			en: "Admin policies",
			de: "Admin-Richtlinien"
		},
		description: {
			en: "Required second factor, session length, allowed modules.",
			de: "Pflicht zum zweiten Faktor, Sitzungsdauer, erlaubte Module."
		}
	},
	{
		id: "audit-log",
		name: {
			en: "Audit log",
			de: "Audit-Log"
		},
		description: {
			en: "Content changes with export and retention.",
			de: "Inhaltsänderungen mit Export und Aufbewahrung."
		}
	},
	{
		id: "team-tabs",
		name: {
			en: "Team tabs",
			de: "Team-Tabs"
		},
		description: {
			en: "Dashboard tabs that admins maintain for groups.",
			de: "Dashboard-Tabs, die Admins für Gruppen pflegen."
		}
	},
	{
		id: "announcements",
		name: {
			en: "Announcements",
			de: "Ankündigungen"
		},
		description: {
			en: "Notices for everyone, with read marks and expiry.",
			de: "Hinweise für alle, mit Gelesen-Markierung und Ablauf."
		}
	},
	{
		id: "absences",
		name: {
			en: "Absences and team calendar",
			de: "Abwesenheiten und Teamkalender"
		},
		description: {
			en: "Who is away when, for the whole team.",
			de: "Wer wann weg ist, für das ganze Team."
		}
	},
	{
		id: "presence",
		name: {
			en: "Live presence",
			de: "Anwesenheit"
		},
		description: {
			en: "Who is in the office, at home or away.",
			de: "Wer im Büro, zu Hause oder unterwegs ist."
		}
	},
	{
		id: "branding",
		name: {
			en: "Branding",
			de: "Branding"
		},
		description: {
			en: "Own name, logo and colours for the instance.",
			de: "Eigener Name, Logo und Farben für die Instanz."
		}
	},
	{
		id: "connected-accounts",
		name: {
			en: "Connected accounts",
			de: "Verbundene Konten"
		},
		description: {
			en: "Each member links their own account at an external service.",
			de: "Jedes Mitglied verbindet sein eigenes Konto bei einem externen Dienst."
		}
	},
	{
		id: "workplace-search",
		name: {
			en: "Workplace search",
			de: "Arbeitsplatz-Suche"
		},
		description: {
			en: "Confluence, SharePoint, Notion, Slack and Teams in the search field.",
			de: "Confluence, SharePoint, Notion, Slack und Teams im Suchfeld."
		}
	},
	{
		id: "chat-apps",
		name: {
			en: "Chat apps",
			de: "Chat-Apps"
		},
		description: {
			en: "The start page in Slack and Microsoft Teams.",
			de: "Die Startseite in Slack und Microsoft Teams."
		}
	}
], Kr = class {
	http;
	constructor(e) {
		this.http = e;
	}
	async levels() {
		return (await this.http.get("/api/licence/features")).features;
	}
	get() {
		return this.http.get("/api/licence");
	}
	install(e) {
		return this.http.put("/api/licence", { key: e.trim() });
	}
	remove() {
		return this.http.delete("/api/licence");
	}
};
function qr(e) {
	return `status_${e.status.replace("-", "_")}`;
}
// src/licence/texts.ts
var Jr = {
	intro: {
		en: "Features for organisations are part of this edition. Each one works once a licence key unlocks it; the others stay listed here.",
		de: "Funktionen für Organisationen gehören zu dieser Ausgabe. Jede funktioniert, sobald ein Lizenzschlüssel sie freischaltet; die übrigen bleiben hier aufgelistet."
	},
	on: {
		en: "On",
		de: "An"
	},
	read: {
		en: "Read only",
		de: "Nur lesen"
	},
	off: {
		en: "Off",
		de: "Aus"
	},
	key_title: {
		en: "Licence key",
		de: "Lizenzschlüssel"
	},
	status_missing: {
		en: "No licence key yet.",
		de: "Noch kein Lizenzschlüssel."
	},
	status_valid: {
		en: "Licensed to {licensee} until {until}.",
		de: "Lizenziert für {licensee} bis {until}."
	},
	status_expired: {
		en: "The key for {licensee} expired on {until}. Data stays readable; changes are paused.",
		de: "Der Schlüssel für {licensee} ist am {until} abgelaufen. Daten bleiben lesbar, Änderungen sind pausiert."
	},
	status_over_seats: {
		en: "{accounts} active accounts, the key covers {seats}. Data stays readable; changes are paused.",
		de: "{accounts} aktive Konten, der Schlüssel deckt {seats}. Daten bleiben lesbar, Änderungen sind pausiert."
	},
	status_invalid: {
		en: "The stored key is not valid: {problem}",
		de: "Der gespeicherte Schlüssel ist ungültig: {problem}"
	},
	seats: {
		en: "Accounts: {accounts} of {seats}",
		de: "Konten: {accounts} von {seats}"
	},
	from_env: {
		en: "The key is set with LICENCE_KEY on the server and can only be changed there.",
		de: "Der Schlüssel ist auf dem Server mit LICENCE_KEY gesetzt und lässt sich nur dort ändern."
	},
	paste: {
		en: "Paste a licence key",
		de: "Lizenzschlüssel einfügen"
	},
	install: {
		en: "Save key",
		de: "Schlüssel speichern"
	},
	remove: {
		en: "Remove key",
		de: "Schlüssel entfernen"
	},
	remove_confirm: {
		en: "Remove the licence key? Licensed features switch off; their data is kept.",
		de: "Lizenzschlüssel entfernen? Lizenzierte Funktionen schalten ab, ihre Daten bleiben erhalten."
	},
	saved: {
		en: "Saved.",
		de: "Gespeichert."
	},
	unknown: {
		en: "The key also unlocks features this version does not have yet: {list}",
		de: "Der Schlüssel schaltet auch Funktionen frei, die diese Version noch nicht hat: {list}"
	},
	admins_only: {
		en: "Only admins can see and change the licence key.",
		de: "Nur Admins können den Lizenzschlüssel sehen und ändern."
	}
}, Yr = /* @__PURE__ */ sr("<li><div class=\"svelte-1km193d\"><strong class=\"svelte-1km193d\"> </strong> <span class=\"muted small svelte-1km193d\"> </span></div> <span> </span></li>"), Xr = /* @__PURE__ */ sr("<p class=\"muted\"> </p>"), Zr = /* @__PURE__ */ sr("<p class=\"muted small svelte-1km193d\"> </p>"), Qr = /* @__PURE__ */ sr("<button class=\"btn btn-danger\" type=\"button\"> </button>"), $r = /* @__PURE__ */ sr("<form class=\"svelte-1km193d\"><textarea class=\"input svelte-1km193d\" rows=\"4\" required=\"\"></textarea> <div class=\"actions svelte-1km193d\"><button class=\"btn btn-primary\" type=\"submit\"> </button> <!></div></form>"), ei = /* @__PURE__ */ sr("<p role=\"status\"> </p> <!> <!> <!>", 1), ti = /* @__PURE__ */ sr("<p class=\"error\" role=\"alert\"> </p>"), ni = /* @__PURE__ */ sr("<p class=\"muted\" role=\"status\"> </p>"), ri = /* @__PURE__ */ sr("<p class=\"muted\"> </p> <ul class=\"features svelte-1km193d\"></ul> <h3 class=\"svelte-1km193d\"> </h3> <!> <!> <!>", 1), ii = {
	hash: "svelte-1km193d",
	code: ".features.svelte-1km193d {list-style:none;padding:0;margin:0 0 var(--space-3);}.features.svelte-1km193d li:where(.svelte-1km193d) {display:flex;gap:var(--space-2);justify-content:space-between;align-items:center;padding:var(--space-2) 0;border-bottom:1px solid var(--color-border, rgba(127, 127, 127, 0.2));}.features.svelte-1km193d li:where(.svelte-1km193d) div:where(.svelte-1km193d) {display:flex;flex-direction:column;}.features.svelte-1km193d li.off:where(.svelte-1km193d) strong:where(.svelte-1km193d) {opacity:0.7;}.small.svelte-1km193d {font-size:0.85em;}.state.svelte-1km193d {font-size:0.85em;padding:0 var(--space-2);border-radius:999px;white-space:nowrap;}.state-full.svelte-1km193d {background:var(--color-success-bg, rgba(34, 197, 94, 0.15));}.state-read.svelte-1km193d {background:var(--color-warning-bg, rgba(234, 179, 8, 0.18));}.state-none.svelte-1km193d {background:var(--color-surface-2, rgba(127, 127, 127, 0.12));}form.svelte-1km193d {display:grid;gap:var(--space-2);}textarea.svelte-1km193d {font-family:var(--font-mono, monospace);font-size:0.85em;}.actions.svelte-1km193d {display:flex;gap:var(--space-2);}h3.svelte-1km193d {font-size:1rem;margin:var(--space-3) 0 var(--space-2);}"
};
function ai(e, t) {
	Be(t, !0), Ar(e, ii);
	let n = /* @__PURE__ */ lt(() => new Kr(t.context.http)), r = (e, n = {}) => Wr(Jr, e, t.context.language, n), i = (e) => e[t.context.language] ?? e.en, a = /* @__PURE__ */ M(Ut({})), o = /* @__PURE__ */ M(null), s = /* @__PURE__ */ M(""), c = /* @__PURE__ */ M(""), l = /* @__PURE__ */ M(""), u = /* @__PURE__ */ M(!1);
	async function d() {
		try {
			N(a, await Z(n).levels(), !0), t.context.account.admin && N(o, await Z(n).get(), !0);
		} catch (e) {
			N(c, t.context.errorMessage(e), !0);
		}
	}
	ln(() => {
		d();
	});
	async function f(e) {
		e.preventDefault(), N(c, N(l, ""), !0), N(u, !0);
		try {
			N(o, await Z(n).install(Z(s)), !0), N(s, ""), N(l, r("saved"), !0), N(a, await Z(n).levels(), !0);
		} catch (e) {
			N(c, t.context.errorMessage(e), !0);
		} finally {
			N(u, !1);
		}
	}
	async function p() {
		if (confirm(r("remove_confirm"))) {
			N(c, N(l, ""), !0);
			try {
				await Z(n).remove(), await d();
			} catch (e) {
				N(c, t.context.errorMessage(e), !0);
			}
		}
	}
	let m = (e) => r(e === "full" ? "on" : e === "read" ? "read" : "off");
	var h = ri(), g = Qt(h), _ = F(g, !0), v = I(g, 2);
	wr(v, 21, () => Gr, (e) => e.id, (e, t) => {
		let n = /* @__PURE__ */ lt(() => Z(a)[Z(t).id]);
		var r = Yr();
		let o;
		var s = Zt(r), c = Zt(s), l = F(c, !0), u = F(I(c, 2), !0);
		Se(s);
		var d = I(s, 2), f = F(d, !0);
		Se(r), z((e, t, i) => {
			o = Nr(r, 1, "svelte-1km193d", null, o, { off: !Z(n) || Z(n) === "none" }), $(l, e), $(u, t), Nr(d, 1, `state state-${Z(n) ?? "none" ?? ""}`, "svelte-1km193d"), $(f, i);
		}, [
			() => i(Z(t).name),
			() => i(Z(t).description),
			() => m(Z(n))
		]), Q(e, r);
	}), Se(v);
	var y = I(v, 2), b = F(y, !0), x = I(y, 2), S = (e) => {
		var t = Xr(), n = F(t, !0);
		z((e) => $(n, e), [() => r("admins_only")]), Q(e, t);
	}, ee = (e) => {
		var t = ei(), n = Qt(t), i = F(n, !0), a = I(n, 2), c = (e) => {
			var t = Zr(), n = F(t, !0);
			z((e) => $(n, e), [() => r("seats", {
				accounts: Z(o).accounts,
				seats: Z(o).seats
			})]), Q(e, t);
		};
		br(a, (e) => {
			Z(o).status !== "missing" && Z(o).seats > 0 && e(c);
		});
		var l = I(a, 2), d = (e) => {
			var t = Zr(), n = F(t, !0);
			z((e) => $(n, e), [() => r("unknown", { list: Z(o).unknown.join(", ") })]), Q(e, t);
		};
		br(l, (e) => {
			Z(o).unknown.length && e(d);
		});
		var m = I(l, 2), h = (e) => {
			var t = Xr(), n = F(t, !0);
			z((e) => $(n, e), [() => r("from_env")]), Q(e, t);
		}, g = (e) => {
			var t = $r(), n = Zt(t);
			Ze(n);
			var i = I(n, 2), a = Zt(i), c = F(a, !0), l = I(a, 2), d = (e) => {
				var t = Qr(), n = F(t, !0);
				z((e) => $(n, e), [() => r("remove")]), Qn("click", t, p), Q(e, t);
			};
			br(l, (e) => {
				Z(o).source === "settings" && e(d);
			}), Se(i), Se(t), z((e, t, r) => {
				Lr(n, "placeholder", e), Lr(n, "aria-label", t), a.disabled = Z(u), $(c, r);
			}, [
				() => r("paste"),
				() => r("paste"),
				() => r("install")
			]), Zn("submit", t, f), Vr(n, () => Z(s), (e) => N(s, e)), Q(e, t);
		};
		br(m, (e) => {
			Z(o).source === "environment" ? e(h) : e(g, -1);
		}), z((e) => $(i, e), [() => r(qr(Z(o)), {
			licensee: Z(o).licensee ?? "",
			until: Z(o).until ?? "",
			seats: Z(o).seats,
			accounts: Z(o).accounts,
			problem: Z(o).problem ?? ""
		})]), Q(e, t);
	};
	br(x, (e) => {
		t.context.account.admin ? Z(o) && e(ee, 1) : e(S);
	});
	var C = I(x, 2), te = (e) => {
		var t = ti(), n = F(t, !0);
		z(() => $(n, Z(c))), Q(e, t);
	};
	br(C, (e) => {
		Z(c) && e(te);
	});
	var ne = I(C, 2), re = (e) => {
		var t = ni(), n = F(t, !0);
		z(() => $(n, Z(l))), Q(e, t);
	};
	br(ne, (e) => {
		Z(l) && e(re);
	}), z((e, t) => {
		$(_, e), $(b, t);
	}, [() => r("intro"), () => r("key_title")]), Q(e, h), Ve();
}
$n(["click"]);
// src/lib/mountable.ts
function oi(e) {
	return { mount(t, n, r) {
		let i = mr(e, {
			target: t,
			props: {
				...n,
				context: r
			}
		});
		return () => void vr(i);
	} };
}
// src/plugin.ts
var si = {
	apiVersion: 1,
	name: "ee",
	widgets: [],
	settings: [{
		id: "licence",
		label: {
			en: "Features for organisations",
			de: "Funktionen für Organisationen"
		},
		description: {
			en: "",
			de: ""
		},
		view: oi(ai)
	}]
};
export { si as plugin };
