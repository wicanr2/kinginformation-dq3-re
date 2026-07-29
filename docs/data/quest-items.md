# DQ3 劇情道具 id 對照表(RE 定錨)

> 來源:font 渲染 D3TXT00(rec = item_code + 1)+ handler 反組譯交叉確認。
> 杜勝利攻略各鏈接線查表用。劇情物多為**連續 id**(0x59-0x6b 一整段)。

## 劇情關鍵物

| id | 名稱 | 取得 / 用途(杜勝利章)| remake 狀態 |
|---|---|---|---|
| 0x0f | 蓋亞之劍 | CTY55 牢獄祠堂寶箱;用→開火山往尼羅肯特(Ch40)| ✅ 寶箱 + DQ3_USE_GAIA |
| 0x10 | 誘惑之劍 | CTY16 答謝 NPC(byte4=84)| ✅ |
| 0x14 | 草薙之劍 | 八頭大蛇洞窟盡頭(怪75 戰前/掉)| ✅ 八頭大蛇勝利分支給(commit 0e859a0)|
| 0x33 | 金皇冠 | 香巴尼塔甘達特（Ch6）；還羅馬利亞王 | ✅ CTY10 handler14 勝利+接受求饒→清原版 flag0x2e→sec5 `(4,4)` 寶箱放行；見 docs/85 |
| 0x4b | 水槍 | CTY22 byte4=31 | ✅ |
| 0x55 | 盜賊鑰匙 | 拿吉米之塔(byte4=12);門 tier1 | ✅ |
| 0x56 | 魔法鑰匙 | 金字塔 CTY13 寶箱(Ch14);門 tier2 | ✅ 寶箱 |
| 0x57 | 最終鑰匙 | 四島礁 CTY40 寶箱(Ch27);門 tier3 | ✅ 寶箱 + 乾渴壺吸海 |
| 0x58 | 魔法玉(球)| 雷貝 byte4=7 | ✅ |
| 0x59 | 夢幻紅寶石 | 地底湖 CTY11 寶箱(Ch9)→ 換覺醒粉 | ✅ |
| 0x5a | 覺醒粉 | 精靈女王換(byte4=16);諾阿尼魯解催眠(Ch11)| ✅ transform + DQ3_USE_AWAKEN |
| 0x5b | **國王的信** | 波魯多加王首訪授予→諾魯德 CTY62；handler50 只檢查、不消耗，NPC 引路後由玩家踩 handler57 格開東方通道。船是 vehicle，非 item | ✅ 正式 InputState trace + save/load（docs/88） |
| 0x5c | 黑胡椒 | 巴哈拉達古布達 handler25；完成 CTY14 救援後，flag0x36 set 時一次性取得 | ✅ `hostage_rescue` 原版 parity、正式 InputState trace、runtime 圖與 save/load（docs/89） |
| 0x5d | 隱身草 | 朗錫爾道具店(Ch25);耶進貝亞隱身 | ✅ 商店貨架 |
| 0x5e | 乾渴壺 | 耶進貝亞推石密道(Ch26);吸海(Ch27)| ✅ DQ3_USE_DRAIN(取得鏈待)|
| 0x5f | 黑暗之燈 | 提頓村武器店 2F(Ch22);變夜 | ✅ 寶箱(CTY20 sec1,flag0x39)|
| 0x60 | 回音之笛 | 亞布之塔 3F(Ch31)| ✅ 寶箱(CTY26 sec2,flag0xb0)|
| 0x61 | 拉之鏡 | 拉之鏡洞窟(Ch35);沙曼歐莎假國王現形 | ✅ 寶箱(CTY24 sec2,flag0x9f)|
| 0x62 | 變身杖 | 怪力魔(怪89)掉(Ch35)→ 換船員之骨 | ✅ boss + transform(byte4=44)|
| 0x63 | 船員之骨 | 雪地草原換(byte4=44,Ch37);找幽靈船(Ch38)| ✅ |
| 0x64 | 愛的回憶 | 幽靈船 CTY36 寶箱(Ch38);歐里比雅解詛咒(Ch39)| ✅ 寶箱 |
| 0x65 | 光之珠 | 龍王城龍之女王(byte4=52)| ✅ |
| 0x66 | 綠寶珠 | 提頓村牢房(Ch28)| ✅ |
| 0x67 | 藍寶珠 | 勇氣洞窟 CTY23 寶箱(Ch30)| ✅ |
| 0x68 | 紅寶珠 | 海盜村 CTY27 密道寶箱(Ch32)| ✅ |
| 0x69 | 紫寶珠 | 八頭大蛇戰勝(Ch24)| ✅ |
| 0x6a | 黃寶珠 | 新城鎮 CTY83 椅後(Ch41)| ✅ |
| 0x6b | 銀寶珠 | 尼羅肯特祠堂(byte4=49,Ch40)| ✅ |
| 0x72 | 太陽之石 | 下層拉達多姆/CTY80(Ch47);彩虹材料 | ✅ 寶箱(CTY80 sec2,flag0xc2)|
| 0x73 | 雲雨之杖 | 精靈祠堂 CTY92(持精靈的守護換)| ✅ transform |
| 0x74 | 精靈的守護 | 魯比斯之塔 CTY82 用妖精之笛解詛咒得 | ✅ DQ3_USE_FAIRYFLUTE |
| 0x77 | 妖精之笛 | 瑪依拉 CTY81 寶箱(Ch52)→ 魯比斯之塔解詛咒 | ✅ 寶箱 |
| 0x6d | 歐里空金屬 | 達姆杜拉 CTY84(站 23,36 朝上 examine)→ 瑪依拉賣換王者之劍 | ✅ 寶箱(王者之劍條件商店待)|
| 0x70 | 生命戒指 | 利姆達爾旅館(Ch51)| ✅ 寶箱(CTY86 sec0,flag0xd4)|
| 0x71 | 銀豎琴 | 加萊祠堂(Ch49)| ✅ 寶箱(CTY78 sec1,flag0x9a)|
| 0x75 | 彩虹水滴 | 神聖祠堂合成(Ch55,bug#2 修正產 0x75 非銀寶珠);下層用→架彩虹橋通終盤 | ✅ 合成+架橋 |

## 不死鳥 questline(六寶珠 0x66-0x6b)

六珠逐顆放上 CTY70 sec0 六座祭壇 → 與守護神交談 → 六幀蛋動畫 → 拉米亞停在
地表 `(48,189)` → 走上搭乘、Confirm 合法降落。✅ Ebiten R-3 事件切片完成。

> 舊 C remake 的「集滿自動復活」、CTY82 與 `y` 起飛／降落都是推測性簡化，非原版流程。

> remake 接線機制:寶箱(既有 chest handler ep2<0x90)/ sub2 NPC 表(dq3_scripted,含 transform
> consume_prereq)/ 位置 item-use(DQ3_USE_AWAKEN/GAIA/DRAIN)/ boss 事件(boss:<id>[:獎勵])。
