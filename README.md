# bundle-helper

Goで書かれた、C++ファイルのコメントアウト制御ツール。  
※注意：開発中なので、デバッグ用のログがstderrに出まくっています。  

競技プログラミングで用いられるbundleツールに`oj-bundle`がありますが、 `#ifdef` ディレクティブ内に`#include`が含まれていたり `#pragma` が含まれていたりすると正常にincludeの展開が出来ないので、このツールを作成しました。  
`oj-bundle` 実行前にそのようなコードをコメントアウトしておいて、正常にbundleが実行してから、コメントアウトを戻すといった操作が将来的にできるようになる予定です。

## Usage

```bash
go run cmd/main.go /path/to/source.cpp
# /path/to/source.cpp.converted.cpp にコメントアウトが制御されたコードが出力されます
```

### コメントアウトの制御方法

`// @bundle-helper COMMAND` でコメントアウトを制御できます  

これが  

```cpp
// @bundle-helper comment_single_line
constexpr int willBeCommentedOut = 1;

// @bundle-helper uncomment_single_line
// constexpr int willBeUncommented = 1;

// @bundle-helper comment_block_begin
Class WillBeCommentedOut {
    int member;
};
// @bundle-helper comment_block_end

// @bundle-helper uncomment_block_begin
// Class WillBeUncommented {
//     int member;
// };
// @bundle-helper uncomment_block_end
```

↓こうなる  

```cpp
// @bundle-helper comment_single_line
// constexpr int willBeCommentedOut = 1;

// @bundle-helper uncomment_single_line
constexpr int willBeUncommented = 1;

// @bundle-helper comment_block_begin
// Class WillBeCommentedOut {
//     int member;
// };
// @bundle-helper comment_block_end

// @bundle-helper uncomment_block_begin
Class WillBeUncommented {
    int member;
};
// @bundle-helper uncomment_block_end
```
