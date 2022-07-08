(def ch (make-chan))

; test that channels and symbol translation are working
(go (send! ch 'foo))
(assert (= 'foo (<! ch)))

; test that coroutines share the same global scope
(def global "foo")
(go (send! ch '()) (def global "bar"))
(<! ch)
(assert (= global "bar"))
