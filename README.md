# 🧩 MySQL → Oracle Custom Driver & SQL Rewriter

This project provides a **custom database driver and SQL rewriter** that enables **GORM** applications originally built for **MySQL** to run seamlessly on **Oracle Database** using the **godror** driver.  

It rewrites MySQL-specific SQL syntax into Oracle-compatible SQL dynamically before query execution.

---

## 🚀 Overview

This library serves as a **compatibility layer** between MySQL-based ORM logic and Oracle DB by intercepting and transforming SQL queries at runtime.

**Core Features:**
- ✨ Dynamic SQL rewrite from MySQL syntax to Oracle syntax  
- ⚙️ Plug-and-play with existing GORM-based repositories  
- 🔄 Uses `godror` as the underlying Oracle driver  
- 🧠 Supports DML operations: `SELECT`, `INSERT`, `UPDATE`, `DELETE`  
- 🔍 Optional SQL logging and debugging hooks  
