// 切换到im数据库
use im_db;

// 创建消息集合及索引（去重）
if (!db.getCollectionNames().includes('messages')) {
  db.createCollection('messages');
}
db.messages.createIndex({ conversation_id: 1, send_time: -1 });
db.messages.createIndex({ receiver_id: 1, status: 1, send_time: -1 });
db.messages.createIndex({ sender_id: 1, send_time: -1 });
db.messages.createIndex({ message_id: 1 }, { unique: true });
db.messages.createIndex({ is_read: 1 });

// 创建用户状态集合及索引
if (!db.getCollectionNames().includes('user_status')) {
  db.createCollection('user_status');
}
db.user_status.createIndex({ status: 1 });

// 创建索引集合（用于存储分布式ID生成器的序列）
if (!db.getCollectionNames().includes('sequences')) {
  db.createCollection('sequences');
}
// 初始化消息ID序列
if (db.sequences.countDocuments({ _id: 'message_id' }) === 0) {
  db.sequences.insertOne({ _id: 'message_id', value: 0 });
}

// 示例消息插入
if (db.messages.countDocuments({ conversation_id: 10001 }) === 0) {
  db.messages.insertOne({
    conversation_id: 10001,
    sender_id: 10002,
    content: 'hello world',
    msg_type: 1, // 1文本2图片3文件...
    send_time: new Date().getTime(),
    is_read: false,
    is_recalled: false,
    extra: {}
  });
}

print('MongoDB initialization completed successfully');