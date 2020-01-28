package jps

import "time"

type CacheMode uint8

const (
	//CacheModeUnspecified is a mode which leaves the cache settings unspecified.
	//Some protocols may not support this, in which case, CacheModeMustRevalidate is used.
	CacheModeUnspecified CacheMode = iota
	//CacheModeNoCache
	CacheModeNoCache
	CacheModeMustRevalidate
	CacheModeNoStore
	CacheModeNoTransform

	/*
		NoStore
		AlwaysRevalidate
		MustRevalidate
	*/
)

/*
 Response directives:
  must-revalidate
   The "must-revalidate" response directive indicates that once it has
   become stale, a cache MUST NOT use the response to satisfy subsequent
   requests without successful validation on the origin server.

   The must-revalidate directive is necessary to support reliable
   operation for certain protocol features.  In all circumstances a
   cache MUST obey the must-revalidate directive; in particular, if a
   cache cannot reach the origin server for any reason, it MUST generate
   a 504 (Gateway Timeout) response.

   The must-revalidate directive ought to be used by servers if and only
   if failure to validate a request on the representation could result
   in incorrect operation, such as a silently unexecuted financial
   transaction.

  no-cache
   The "no-cache" response directive indicates that the response MUST
   NOT be used to satisfy a subsequent request without successful
   validation on the origin server.  This allows an origin server to
   prevent a cache from using it to satisfy a request without contacting
   it, even by caches that have been configured to send stale responses.

   If the no-cache response directive specifies one or more field-names,
   then a cache MAY use the response to satisfy a subsequent request,
   subject to any other restrictions on caching.  However, any header
   fields in the response that have the field-name(s) listed MUST NOT be
   sent in the response to a subsequent request without successful
   revalidation with the origin server.  This allows an origin server to
   prevent the re-use of certain header fields in a response, while
   still allowing caching of the rest of the response.

   The field-names given are not limited to the set of header fields
   defined by this specification.  Field names are case-insensitive.
   This directive uses the quoted-string form of the argument syntax.  A
   sender SHOULD NOT generate the token form (even if quoting appears
   not to be needed for single-entry lists).

   Note: Although it has been back-ported to many implementations, some
   HTTP/1.0 caches will not recognize or obey this directive.  Also,
   no-cache response directives with field-names are often handled by
   caches as if an unqualified no-cache directive was received; i.e.,
   the special handling for the qualified form is not widely
   implemented.

  no-store
   The "no-store" response directive indicates that a cache MUST NOT
   store any part of either the immediate request or response.  This
   directive applies to both private and shared caches.  "MUST NOT
   store" in this context means that the cache MUST NOT intentionally
   store the information in non-volatile storage, and MUST make a
   best-effort attempt to remove the information from volatile storage
   as promptly as possible after forwarding it.

   This directive is NOT a reliable or sufficient mechanism for ensuring
   privacy.  In particular, malicious or compromised caches might not
   recognize or obey this directive, and communications networks might
   be vulnerable to eavesdropping.
*/

/*
 Request directives:
  no-cache
   The "no-cache" request directive indicates that a cache MUST NOT use
   a stored response to satisfy the request without successful
   validation on the origin server.

  no-store
   The "no-store" request directive indicates that a cache MUST NOT
   store any part of either this request or any response to it.  This
   directive applies to both private and shared caches.  "MUST NOT
   store" in this context means that the cache MUST NOT intentionally
   store the information in non-volatile storage, and MUST make a
   best-effort attempt to remove the information from volatile storage
   as promptly as possible after forwarding it.

   This directive is NOT a reliable or sufficient mechanism for ensuring
   privacy.  In particular, malicious or compromised caches might not
   recognize or obey this directive, and communications networks might
   be vulnerable to eavesdropping.

   Note that if a request containing this directive is satisfied from a
   cache, the no-store request directive does not apply to the already
   stored response.

  only-if-cached
   The "only-if-cached" request directive indicates that the client only
   wishes to obtain a stored response.  If it receives this directive, a
   cache SHOULD either respond using a stored response that is
   consistent with the other constraints of the request, or respond with
   a 504 (Gateway Timeout) status code.  If a group of caches is being
   operated as a unified system with good internal connectivity, a
   member cache MAY forward such a request within that group of caches.
*/

type CacheType uint8

const (
	CacheTypePublic CacheType = 2 << iota
	CacheTypePrivate
	CacheTypeAll = CacheTypePublic | CacheTypePrivate
)

func (t CacheType) AppliesToPublic() bool {
	return t&CacheTypePublic != 0
}

func (t CacheType) AppliesToPrivate() bool {
	return t&CacheTypePublic != 0
}

//ResponseCacheSettingsGetter specifies the settings that are required for a cache.
type ResponseCacheSettingsGetter interface {
	//CacheSupported returns true if the underlying protocol supports specifying cache properties, or false otherwise.
	CacheSupported() bool

	CacheMode() (CacheMode, error)
	CachePublic() (bool, error)
	CacheMaxAge() (int64, error)
	CacheExpires() (time.Time, error)
	CacheTransformAllowed() (bool, error)
}

type ResponseCacheSettingsSetter interface {
	SetCacheMode(CacheMode) error
	SetCachePublic(bool) error
	SetCacheMaxAge(int64) error
	SetCacheExpires(time.Time) error
	SetCacheTransformAllowed(bool) error
}

type ResponseCacheSettings interface {
	ResponseCacheSettingsGetter
	ResponseCacheSettingsSetter
}

type RequestCacheSettingsGetter interface {
	//CacheSupported returns true if the underlying protocol supports specifying cache properties, or false otherwise.
	CacheSupported() bool

	MaxAge() (int64, error)
	MaxStale() (int64, error)
	MinFresh() (int64, error)
	CacheTransformAllowed() (bool, error)
}

type RequestCacheSettingsSetter interface {
	SetMaxAge(int64) error
	SetMaxStale(int64) error
	SetMinFresh(int64) error
	SetCacheTransformAllowed(bool) error
}

type RequestCacheSettings interface {
	RequestCacheSettingsGetter
	RequestCacheSettingsSetter
}
