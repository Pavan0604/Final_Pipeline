const express = require("express");
const mysql = require("mysql");
const app = express();

app.get("/", function (req, res) {
  res.send("Hello from server V2");
});

app.get("/createTable", function (req, res) {
  let sql =
    "create table employees(id int AUTO_INCREMENT,name VARCHAR(255),designation VARCHAR(255),PRIMARY KEY (id))";
  db.query(sql, (err) => {
    if (err) {
      console.log("err in creating table...", err);
      res.send("error in creating table");
    } else {
      res.send("employee table created");
    }
  });
});

app.get("/add", function (req, res) {
  let post = { name: "pavan", designation: "devops Engineer" };
  let sql = "insert into employees SET ?";
  let query = db.query(sql, post, (err) => {
    if (err) {
      console.log("error in adding a employee record..", err);
      res.send("error in created record");
    } else {
      res.send("Employee record added..");
    }
  });
});

app.get("/get-list", function (req, res) {
  let sql = "select * from employees";
  let query = db.query(sql, (err, result) => {
    if (err) {
      console.log("error in retrieving employee records..", err);
      res.send("error in retrieving employee list");
    } else {
      console.log("results", result);
      res.send("Employee retrived..");
    }
  });
});

const db = mysql.createConnection({
  host:"database-1.crnzykf3tngp.ap-south-1.rds.amazonaws.com",
  user: "root",
  password:"Pavan1234",
  port: "3306",
  database:"employee",
});

db.connect((err) => {
  if (err) {
    console.log("connecting to db error", err);
  } else {
    console.log("mysql connected");
  }
});

app.listen(5000, function () {
  console.log("App listening on port 5000");
});
