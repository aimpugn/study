<?php

require_once dirname(dirname(__DIR__)) . "/vendor/autoload.php";

use Finance\Chai\Gateway\Transaction\CheckoutCardForm;

$json =<<<EOF
{
  "form": {
    "form": {
      "store_id": "store_id",
      "payment_id": "payment_id",
      "order_name": "order_name",
      "is_cultural_expense": false,
      "is_escrow": false,
      "escrow_products": [
        {
          "id": "id",
          "name": "name",
          "code": "code",
          "unitPrice": 0,
          "quantity": 0
        }
      ],
      "customer_form": {
        "customer_id": "customer_id",
        "phone_number": "phone_number",
        "customer_name": {
          "full_name": "full_name",
          "separated_name": {
            "first_name": "first_name",
            "last_name": "last_name"
          }
        },
        "email": "email",
        "zipcode": "zipcode",
        "address": "address"
      },
      "custom_data": "hello world",
      "total_amount": {
        "units": "0",
        "nanos": 0
      },
      "tax_free_amount": {
        "units": "0",
        "nanos": 0
      },
      "currency": "CURRENCY_KRW",
      "origin": {
        "user_agent": "user_agent",
        "url": "url",
        "ip_address": "ip_address",
        "platform_type": "PLATFORM_TYPE_PC"
      },
      "notice_url": "notice_url",
      "confirm_url": "confirm_url"
    },
    "app_scheme": "app_scheme",
    "redirect_url": "redirect_url",
    "channel_selector": {
      "id": {
        "id": "id"
      },
      "condition": {
        "pg_provider": "PG_PROVIDER_DANAL",
        "is_test_channel": false
      }
    },
    "payment_method_form": {
      "card": {
        "card_company": "CARD_COMPANY_BC",
        "installment": {
          "fixed_month": 0,
          "available_month_list": [
            0
          ]
        },
        "use_card_point": false
      },
      "virtualAccount": {
        "receipt_type": "CUSTOMER_TYPE_PERSONAL",
        "account_type": "VIRTUAL_ACCOUNT_TYPE_NORMAL",
        "account_key": "account_key",
        "account_expiry": {
          "valid_hours": 1,
          "due_date": "due_date"
        }
      },
      "transfer": {
        "receipt_type": "CUSTOMER_TYPE_PERSONAL",
        "bank_code": "BANK_TOSS_BANK"
      },
      "mobile": {
        "carrier": "CARRIER_SKT"
      },
      "giftCertificate": {
        "gift_certificate_type": "GIFT_CERTIFICATE_TYPE_BOOKNLIFE"
      },
      "easyPay": {
        "easy_pay_provider": "EASY_PAY_PROVIDER_PAYCO"
      }
    }
  }
}
EOF;

$dummy1 = json_decode($json);
$tmp = $dummy1->test ?: 'test?????';
print_r(['$tmp' => $tmp]);
$tmp1 = $dummy1->form ?: 'no form';
print_r(['$tmp1' => $tmp1]);

//$obj = json_decode($json);
//
//print_r($obj->undefined ?: 'null');
//print_r($obj->form->form->store_id ?: 'null');

exit;