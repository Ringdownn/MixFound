#!/bin/bash

# 更新Go模块导入路径的脚本

# 定义旧模块名和新模块名
OLD_MODULE="MixFound"
NEW_MODULE="MixFound/services/search-engine"

# 更新 go.mod 文件
sed -i '' "s|module $OLD_MODULE|module $NEW_MODULE|g" services/search-engine/go.mod

# 更新所有 Go 文件中的导入路径
find services/search-engine -name "*.go" -type f -exec sed -i '' "s|$OLD_MODULE|$NEW_MODULE|g" {} +

echo "导入路径更新完成！"
