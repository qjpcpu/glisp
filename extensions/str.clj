(defn str/first [& strs]
  "Usage: (str/first & strs)
Return first non empty str or nil."
  (or  (->> (stream strs)
            (reject #(or (not (string? %)) (empty? %)))
            (take 1)
            (realize)
            (car)) ""))

(defn str/join2 [sep elems]
  "Usage: (str/join2 sep elems)

Concatenates the elements of its first argument to create a single string. The separator
string sep is placed between elements in the resulting string."
  (str/join elems sep))
