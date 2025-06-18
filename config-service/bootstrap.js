const { Bootstrap } = require('@midwayjs/bootstrap');
Bootstrap.run().then(() => {
  console.log('Your application is running at http://localhost:7001');
});