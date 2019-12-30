# tg

## 说明
使用类似注解的方式来处理

## Model

### Gen 
* // @tg                // 全部方法 Create Update Info List Delete
* // @tg Create:nosave  // Create时不自动保存
* // @tg Create:nosave;preload=v>V,a>A  // Create时不自动保存，并且预加载数据
* // @tg -Info          // 不需要 Info
* // @tg security=AppUser // 所有的接口都有Security AppUser

注解都是按顺序执行

#### Gen Options

* nosave 不自动保存
* save 自动保存
* preload 预加载  preload=v>V;a>A
* security 权限 security=AppUser,AppKey
* desc 接口组注释

### dbindex
用于更新删除时的主键

### swagger

* maxlength
* minlength
* max
* min
* enums
* params                  // 是否提供接口参数 cu 创建和更新 c仅创建 u仅更新 如果是大写为必填
* pt                      // 自定义swag params type


## Func

### Func Type
* CreateBefore          // 创建执行前
* CreateTxBefore        // 创建事务前
* CreateTxAfter         // 创建事务后提交前
* CreateAfter           // 创建事务执行后
* UpdateBefore
* UpdateTxBefore
* UpdateTxAfter
* UpdateAfter
* InfoBefore
* InfoAfter
* ListBefore
* ListAfter
* DeleteBefore
* DeleteTxBefore
* DeleteTxAfter
* DeleteAfter

### Gen

* // @tg CreateBefore      // 注册在所有Model `CreateBefore`
* // @tg CreateBefore@99   // 注册在所有Model `CreateBefore`优先级为99 数越大 优先级越高，数相等随机
* // @tg CreateBefore:User -UpdateBefor:User    // 注册在`User`的`CreateBefore` 注册在除了`User`的`UpdateBefor`
* // @tg CreateBefore:User@99 -UpdateBefor:User // 注册在`User`的`CreateBefore`且优先级为99 注册在除了`User`的`UpdateBefor`
