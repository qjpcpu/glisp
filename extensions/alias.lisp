(defmac alias [new old]
 `(defn ~new [& body] (apply ~old body))
)
