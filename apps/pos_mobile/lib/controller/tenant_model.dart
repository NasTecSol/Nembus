typedef Tenant = Data;

class TenantModel {
  int? statusCode;
  String? message;
  List<Data>? data;

  TenantModel({this.statusCode, this.message, this.data});

  TenantModel.fromJson(Map<String, dynamic> json) {
    statusCode = json["statusCode"] is int
        ? json["statusCode"]
        : int.tryParse(json["statusCode"]?.toString() ?? '');
    message = json["message"]?.toString();

    final rawData = json["data"];
    if (rawData is List) {
      data = rawData
          .whereType<Map<String, dynamic>>()
          .map((e) => Data.fromJson(e))
          .toList();
    } else if (rawData is Map<String, dynamic>) {
      data = [Data.fromJson(rawData)];
    } else if (json["id"] != null || json["tenant_name"] != null || json["slug"] != null) {
      data = [Data.fromJson(json)];
    } else {
      data = null;
    }
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> dataMap = <String, dynamic>{};
    dataMap["statusCode"] = statusCode;
    dataMap["message"] = message;
    if (data != null) {
      dataMap["data"] = data?.map((e) => e.toJson()).toList();
    }
    return dataMap;
  }
}

class Data {
  String? id;
  String? tenantName;
  String? slug;
  String? dbConnStr;
  bool? isActive;
  Settings? settings;
  String? createdAt;
  String? updatedAt;

  Data({
    this.id,
    this.tenantName,
    this.slug,
    this.dbConnStr,
    this.isActive,
    this.settings,
    this.createdAt,
    this.updatedAt,
  });

  String get name => tenantName ?? slug ?? 'Unnamed Store';
  String get identifier => slug ?? id ?? '';
  String? get subdomain => slug;

  Data.fromJson(Map<String, dynamic> json) {
    id = json["id"]?.toString() ?? json["tenant_id"]?.toString();
    tenantName = json["tenant_name"]?.toString() ?? json["name"]?.toString();
    slug = json["slug"]?.toString() ?? json["identifier"]?.toString() ?? json["code"]?.toString();
    dbConnStr = json["db_conn_str"]?.toString();
    isActive = json["is_active"] ?? json["active"];
    settings = json["settings"] == null ? null : Settings.fromJson(json["settings"]);
    createdAt = json["created_at"]?.toString();
    updatedAt = json["updated_at"]?.toString();
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> dataMap = <String, dynamic>{};
    dataMap["id"] = id;
    dataMap["tenant_name"] = tenantName;
    dataMap["slug"] = slug;
    dataMap["db_conn_str"] = dbConnStr;
    dataMap["is_active"] = isActive;
    if (settings != null) {
      dataMap["settings"] = settings?.toJson();
    }
    dataMap["created_at"] = createdAt;
    dataMap["updated_at"] = updatedAt;
    return dataMap;
  }
}

class Settings {
  String? plan;
  List<String>? features;
  int? maxUsers;

  Settings({this.plan, this.features, this.maxUsers});

  Settings.fromJson(Map<String, dynamic> json) {
    plan = json["plan"]?.toString();
    features = json["features"] == null ? null : List<String>.from(json["features"]);
    maxUsers = json["max_users"];
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> dataMap = <String, dynamic>{};
    dataMap["plan"] = plan;
    if (features != null) {
      dataMap["features"] = features;
    }
    dataMap["max_users"] = maxUsers;
    return dataMap;
  }
}
