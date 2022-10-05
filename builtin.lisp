;; define global variables
(def nil '())

(defn flatten [x]
  "Usage: (flatten coll)
Takes any nested combination of sequential things (lists, vectors,
etc.) and returns their contents as a single, flat foldable
collection."
  (flatmap (fn [e] e) x))
