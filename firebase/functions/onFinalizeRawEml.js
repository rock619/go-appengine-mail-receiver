const functions = require("firebase-functions");
const admin = require("firebase-admin");

if (!admin.apps.length) admin.initializeApp();
const storage = admin.storage();
const srcBucket = functions.config().bucket.src;
const destBucket = functions.config().bucket.dest;

const copyFile = async (
  srcBucketName,
  srcFilename,
  destBucketName,
  destFilename
) => {
  await storage
    .bucket(srcBucketName)
    .file(srcFilename)
    .copy(storage.bucket(destBucketName).file(destFilename));

  console.log(
    `gs://${srcBucketName}/${srcFilename} copied to gs://${destBucketName}/${destFilename}.`
  );
};

exports.onFinalizeRawEml = functions.region("asia-northeast1").storage
  .bucket(srcBucket)
  .object()
  .onFinalize(async (object) => {
    await copyFile(object.bucket, object.name, destBucket, object.name).catch(
      (err) => {
        console.error(err);
      }
    );
  });
