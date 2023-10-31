# cmod

Extract used only sources from huge c/c++ project to embedded into other project as in-tree dependency.

e.g.:

```bash
cmod -v 4 -p C:/Users/lizhen/Downloads/folly-2023.09.04.00 -i folly/concurrency/ConcurrentHashMap.h
```

This will produce a `dist` folder in the input path, which contains only needed files of wanted `folly/concurrency/ConcurrentHashMap.h`
