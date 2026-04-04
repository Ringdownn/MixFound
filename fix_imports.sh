#!/bin/bash

# 修复导入路径的脚本

# 更新所有 Go 文件中的导入路径
find services/search-engine -name "*.go" -type f -exec sed -i '' 's|MixFound/services/search-engine/searcher|MixFound/services/search-engine/internal/searcher|g' {} +
find services/search-engine -name "*.go" -type f -exec sed -i '' 's|MixFound/services/search-engine/web|MixFound/services/search-engine/internal/web|g' {} +
find services/search-engine -name "*.go" -type f -exec sed -i '' 's|MixFound/services/search-engine/core|MixFound/services/search-engine/internal/core|g' {} +
find services/search-engine -name "*.go" -type f -exec sed -i '' 's|MixFound/services/search-engine/global|MixFound/services/search-engine/internal/global|g' {} +
find services/search-engine -name "*.go" -type f -exec sed -i '' 's|MixFound/services/search-engine/redis|MixFound/services/search-engine/internal/redis|g' {} +

echo "导入路径修复完成！"
