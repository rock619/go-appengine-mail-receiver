const functions = require('firebase-functions');
const simpleParser = require('mailparser').simpleParser;
const admin = require('firebase-admin');

if (!admin.apps.length) admin.initializeApp();
const storage = admin.storage();

const bucket = functions.config().bucket.dest;

const newAttachments = (attachments) =>
  attachments.map((m) => ({
    ...m,
    headers: Object.fromEntries(m.headers),
  }));

const headers = (headers) => Object.fromEntries(headers);

exports.onFinalizeEml = functions
  .region('asia-northeast1')
  .runWith({ memory: '1GB' })
  .storage.bucket(bucket)
  .object()
  .onFinalize(async (object) => {
    if (!object.name.endsWith('.eml')) return;

    const mail = await simpleParser((await storage.bucket(object.bucket).file(object.name).download())[0]);
    mail.type = 'email';
    mail.attachments = newAttachments(mail.attachments);
    mail.headers = headers(mail.headers);

    await storage.bucket(object.bucket).file(object.name.replace('.eml', '.json')).save(JSON.stringify(mail));
  });
