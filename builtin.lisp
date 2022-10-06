;; define global variables
(def nil '())

(defn flatten [x]
  "Usage: (flatten coll)
Takes any nested combination of sequential things (lists, vectors,
etc.) and returns their contents as a single, flat foldable
collection."
  (flatmap (fn [e] e) x))

(defn reject [pred coll]
  "Usage: (reject pred coll)
Returns a sequence of the items in coll for which
(pred item) returns logical false. pred must be free of side-effects.

When coll is hash, pred function should take 2 arguments, which are hash key-value pair.
"
  (filter (fn [e] (not (pred e))) coll))
